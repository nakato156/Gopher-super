package recommend

import (
	"fmt"
	"goflix/worker-node/internal/types"
	"math"
	"sort"
	"strings"
)

func PredictForUserItem(u int, i int, userRatings map[int]map[int]float64, nbrs types.NeighborList) (float64, bool) {
	ru, ok := userRatings[u]
	if !ok {
		return 0, false
	}
	neighbors, ok := nbrs[i]
	if !ok || len(neighbors) == 0 {
		return 0, false
	}
	num := 0.0
	den := 0.0
	for _, nb := range neighbors {
		if r, ok := ru[nb.Item]; ok {
			num += nb.Sim * r
			den += math.Abs(nb.Sim)
		}
	}
	if den == 0 {
		return 0, false
	}
	return num / den, true
}

func RecommendTopN(u int, N int, userRatings map[int]map[int]float64, nbrs types.NeighborList, titles map[int]string) []types.Rec {
	ru := userRatings[u]
	seen := make(map[int]struct{}, len(ru))
	for mid := range ru {
		seen[mid] = struct{}{}
	}

	candidates := make(map[int]struct{})
	// candidatos: vecinos de lo que ya vio
	for mid := range ru {
		for _, nb := range nbrs[mid] {
			if _, ok := seen[nb.Item]; !ok {
				candidates[nb.Item] = struct{}{}
			}
		}
	}

	type scored struct {
		id  int
		val float64
	}
	var scores []scored
	for i := range candidates {
		if p, ok := PredictForUserItem(u, i, userRatings, nbrs); ok {
			scores = append(scores, scored{id: i, val: p})
		}
	}
	sort.Slice(scores, func(a, b int) bool { return scores[a].val > scores[b].val })

	if len(scores) > N {
		scores = scores[:N]
	}
	recs := make([]types.Rec, 0, len(scores))
	for _, s := range scores {
		recs = append(recs, types.Rec{MovieID: s.id, Score: s.val, Title: titles[s.id]})
	}
	return recs
}

func HumanList(recs []types.Rec) string {
	var b strings.Builder
	for i, r := range recs {
		fmt.Fprintf(&b, "%2d. (%.3f) %s [id=%d]\n", i+1, r.Score, safeTitle(r.Title), r.MovieID)
	}
	return b.String()
}

func safeTitle(s string) string {
	if s == "" {
		return "<sin título>"
	}
	return s
}
