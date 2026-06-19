package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// communityRankingTypes maps friendly aliases to the server enum.
var communityRankingTypes = map[string]string{
	"influencer": "INFLUENCER",
	"profit":     "TOP_10_PROFIT_ROSS_AMOUNT",
	"followers":  "TOP_10_FOLLOWING_INCREASE",
}

// CommunityRankingType resolves an alias (or raw enum) to the server value.
func CommunityRankingType(alias string) (string, error) {
	a := strings.ToLower(strings.TrimSpace(alias))
	if a == "" {
		return communityRankingTypes["influencer"], nil
	}
	if v, ok := communityRankingTypes[a]; ok {
		return v, nil
	}
	// allow raw enum passthrough (uppercase)
	upper := strings.ToUpper(strings.TrimSpace(alias))
	for _, v := range communityRankingTypes {
		if v == upper {
			return v, nil
		}
	}
	return "", fmt.Errorf("unknown ranking type %q (use: influencer | profit | followers)", alias)
}

type communityItemRaw struct {
	Description         string  `json:"description"`
	ProfitLossAmountKrw float64 `json:"profitLossAmountKrw"`
	ProfitLossRateKrw   float64 `json:"profitLossRateKrw"`
	FollowingCount      int     `json:"followingCount"`
	FollowingIncrease   int     `json:"followingIncrease"`
	Target              struct {
		Nickname      string `json:"nickname"`
		UserProfileID int64  `json:"userProfileId"`
	} `json:"target"`
	Type string `json:"type"`
}

// GetCommunityRankings returns a community leaderboard (인플루언서·수익률·팔로워
// 증가). 공식 API 에 없는 web 전용 기능.
func (c *Client) GetCommunityRankings(ctx context.Context, rankType string) (domain.CommunityRanking, error) {
	serverType, err := CommunityRankingType(rankType)
	if err != nil {
		return domain.CommunityRanking{}, err
	}
	var envelope quoteEnvelope[struct {
		Items []communityItemRaw `json:"items"`
	}]
	endpoint := c.infoBaseURL + "/api/v1/community/top-rankings/" + serverType
	if err := c.getJSON(ctx, endpoint, &envelope); err != nil {
		return domain.CommunityRanking{}, err
	}
	out := domain.CommunityRanking{Type: serverType, FetchedAt: time.Now().UTC()}
	for i, it := range envelope.Result.Items {
		out.Users = append(out.Users, domain.CommunityUser{
			Rank:              i + 1,
			Nickname:          it.Target.Nickname,
			UserProfileID:     it.Target.UserProfileID,
			Description:       it.Description,
			ProfitAmountKRW:   it.ProfitLossAmountKrw,
			ProfitRate:        it.ProfitLossRateKrw,
			FollowingCount:    it.FollowingCount,
			FollowingIncrease: it.FollowingIncrease,
		})
	}
	return out, nil
}
