package middleware

import (
	"context"
	"net/http"

	"github.com/puppe1990/cais/pkg/cais/session"
)

// UserStats holds gamification chrome for layout templates (meta.Site).
type UserStats struct {
	Level  int
	Points int
	Rank   int
}

// UserStatsLoader loads reputation stats for a signed-in user.
type UserStatsLoader interface {
	LoadStats(userID int64) (level, points, rank int, err error)
}

type statsKey struct{}

// LoadUserStats attaches UserStats to the request when a session exists.
func LoadUserStats(loader UserStatsLoader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if loader != nil {
				if id, ok := session.UserID(r); ok {
					if level, points, rank, err := loader.LoadStats(id); err == nil {
						r = r.WithContext(context.WithValue(r.Context(), statsKey{}, UserStats{
							Level:  level,
							Points: points,
							Rank:   rank,
						}))
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UserStatsFrom returns stats injected by LoadUserStats.
func UserStatsFrom(r *http.Request) (UserStats, bool) {
	stats, ok := r.Context().Value(statsKey{}).(UserStats)
	return stats, ok
}
