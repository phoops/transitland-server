package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

func FindTrips(atx sqlx.Ext, limit *int, after *int, ids []int, where *model.TripFilter) (ents []*model.Trip, err error) {
	active := false
	if where != nil && where.FeedVersionSha1 == nil {
		active = true
	}
	q := TripSelect(limit, after, ids, active, where)
	MustSelect(model.DB, q, &ents)
	return ents, nil
}

func TripSelect(limit *int, after *int, ids []int, active bool, where *model.TripFilter) sq.SelectBuilder {
	qView := sq.StatementBuilder.Select(
		"gtfs_trips.*",
		"current_feeds.id AS feed_id",
		"current_feeds.onestop_id AS feed_onestop_id",
		"feed_versions.sha1 AS feed_version_sha1",
	).
		From("gtfs_trips").
		Join("feed_versions ON feed_versions.id = gtfs_trips.feed_version_id").
		Join("current_feeds ON current_feeds.id = feed_versions.feed_id").
		OrderBy("gtfs_trips.id")
	if active {
		qView = qView.Join("feed_states on feed_states.feed_version_id = gtfs_trips.feed_version_id")
	}
	if where != nil {
		if len(where.RouteOnestopIds) > 0 {
			qView = qView.Join("tl_route_onestop_ids tlros on tlros.route_id = gtfs_trips.route_id")
			qView = qView.Where(sq.Eq{"tlros.onestop_id": where.RouteOnestopIds})
		}
		if where.ServiceDate != nil {
			serviceDate := where.ServiceDate.Time
			qView = qView.JoinClause(`
			join lateral (
				select gc.id
				from gtfs_calendars gc 
				left join gtfs_calendar_dates gcda on gcda.service_id = gc.id and gcda.exception_type = 1 and gcda.date = ?::date
				left join gtfs_calendar_dates gcdb on gcdb.service_id = gc.id and gcdb.exception_type = 2 and gcdb.date = ?::date
				where 
					gc.id = gtfs_trips.service_id 
					AND ((
						gc.start_date <= ?::date AND gc.end_date >= ?::date
						AND (CASE EXTRACT(isodow FROM ?::date)
						WHEN 1 THEN monday = 1
						WHEN 2 THEN tuesday = 1
						WHEN 3 THEN wednesday = 1
						WHEN 4 THEN thursday = 1
						WHEN 5 THEN friday = 1
						WHEN 6 THEN saturday = 1
						WHEN 7 THEN sunday = 1
						END)
					) OR gcda.date IS NOT NULL)
					AND gcdb.date is null
				LIMIT 1
			) gc on true
			`, serviceDate, serviceDate, serviceDate, serviceDate, serviceDate)
		}
	}

	q := sq.StatementBuilder.Select("t.*").FromSelect(qView, "t")
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"t.id": ids})
	}
	if after != nil {
		q = q.Where(sq.Gt{"t.id": *after})
	}
	q = q.Limit(checkLimit(limit))
	if where != nil {
		if where.FeedVersionSha1 != nil {
			q = q.Where(sq.Eq{"feed_version_sha1": *where.FeedVersionSha1})
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"feed_onestop_id": *where.FeedOnestopID})
		}
		if len(where.RouteIds) > 0 {
			q = q.Where(sq.Eq{"route_id": where.RouteIds})
		}
		if where.TripID != nil {
			q = q.Where(sq.Eq{"trip_id": *where.TripID})
		}
	}
	return q
}
