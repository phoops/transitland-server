package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

func FindFeeds(atx sqlx.Ext, limit *int, after *int, ids []int, where *model.FeedFilter) (ents []*model.Feed, err error) {
	MustSelect(model.DB, FeedSelect(limit, after, ids, where), &ents)
	return ents, nil
}

func FeedSelect(limit *int, after *int, ids []int, where *model.FeedFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select("t.*").
		From("current_feeds t").
		OrderBy("t.id asc").
		Limit(checkRange(limit, 0, 10_000)).
		Where(sq.Eq{"deleted_at": nil})
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"t.id": ids})
	}
	if after != nil {
		q = q.Where(sq.Gt{"t.id": *after})
	}
	if where != nil {
		if where.Search != nil && len(*where.Search) > 1 {
			rank, wc := tsQuery(*where.Search)
			q = q.Column(rank).Where(wc)
		}
		if where.OnestopID != nil {
			q = q.Where(sq.Eq{"onestop_id": *where.OnestopID})
		}
		if len(where.Spec) > 0 {
			q = q.Where(sq.Eq{"spec": where.Spec})
		}
		// Tags
		if where.Tags != nil {
			for _, k := range where.Tags.Keys() {
				if v, ok := where.Tags.Get(k); ok {
					if v == "" {
						q = q.Where("feed_tags ?? ?", k)
					} else {
						q = q.Where("feed_tags->>? = ?", k, v)
					}
				}
			}
		}
		// Fetch error
		if v := where.FetchError; v == nil {
			// nothing
		} else if *v {
			q = q.Join("feed_states on feed_states.feed_id = t.id").Where(sq.NotEq{"feed_states.last_fetch_error": ""})
		} else if !*v {
			q = q.Join("feed_states on feed_states.feed_id = t.id").Where(sq.Eq{"feed_states.last_fetch_error": ""})
		}
		// Import import status
		if where.ImportStatus != nil {
			// in_progress must be false to check success and vice-versa
			var checkSuccess bool
			var checkInProgress bool
			check := *where.ImportStatus
			if check == "success" {
				checkSuccess = true
				checkInProgress = false
			} else if check == "error" {
				checkSuccess = false
				checkInProgress = false
			} else if check == "in_progress" {
				checkSuccess = false
				checkInProgress = true
			}
			// This lateral join gets the most recent attempt at a completed feed_version_gtfs_import and checks the status
			q = q.JoinClause(`JOIN LATERAL (select fvi.in_progress, fvi.success from feed_versions fv inner join feed_version_gtfs_imports fvi on fvi.feed_version_id = fv.id WHERE fv.feed_id = t.id ORDER BY fvi.id DESC LIMIT 1) fvicheck ON TRUE`).
				Where(sq.Eq{"fvicheck.success": checkSuccess, "fvicheck.in_progress": checkInProgress})
		}
	}
	return q
}
