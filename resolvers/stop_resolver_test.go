package resolvers

import (
	"testing"
)

func TestStopResolver(t *testing.T) {
	bartStops := []string{"12TH", "16TH", "19TH", "19TH_N", "24TH", "ANTC", "ASHB", "BALB", "BAYF", "CAST", "CIVC", "COLS", "COLM", "CONC", "DALY", "DBRK", "DUBL", "DELN", "PLZA", "EMBR", "FRMT", "FTVL", "GLEN", "HAYW", "LAFY", "LAKE", "MCAR", "MCAR_S", "MLBR", "MONT", "NBRK", "NCON", "OAKL", "ORIN", "PITT", "PCTR", "PHIL", "POWL", "RICH", "ROCK", "SBRN", "SFIA", "SANL", "SHAY", "SSAN", "UCTY", "WCRK", "WARM", "WDUB", "WOAK"}
	caltrainRailStops := []string{"70011", "70012", "70021", "70022", "70031", "70032", "70041", "70042", "70051", "70052", "70061", "70062", "70071", "70072", "70081", "70082", "70091", "70092", "70101", "70102", "70111", "70112", "70121", "70122", "70131", "70132", "70141", "70142", "70151", "70152", "70161", "70162", "70171", "70172", "70191", "70192", "70201", "70202", "70211", "70212", "70221", "70222", "70231", "70232", "70241", "70242", "70251", "70252", "70261", "70262", "70271", "70272", "70281", "70282", "70291", "70292", "70301", "70302", "70311", "70312", "70321", "70322"}
	caltrainBusStops := []string{"777402", "777403"}
	caltrainStops := []string{}
	caltrainStops = append(caltrainStops, caltrainRailStops...)
	caltrainStops = append(caltrainStops, caltrainBusStops...)
	allStops := []string{}
	allStops = append(allStops, bartStops...)
	allStops = append(allStops, caltrainStops...)
	vars := hw{"stop_id": "MCAR"}
	testcases := []testcase{
		{
			"basic",
			`query($feed_version_sha1:String!) { stops(where:{feed_version_sha1:$feed_version_sha1}) { stop_id } }`, // just check BART
			hw{"feed_version_sha1": "e535eb2b3b9ac3ef15d82c56575e914575e732e0"},
			``,
			"stops.#.stop_id",
			bartStops,
		},
		{
			"basic fields",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {onestop_id feed_version_sha1 feed_onestop_id location_type stop_code stop_desc stop_id stop_name stop_timezone stop_url wheelchair_boarding zone_id} }`,
			vars,
			`{"stops":[{"feed_onestop_id":"BA","feed_version_sha1":"e535eb2b3b9ac3ef15d82c56575e914575e732e0","location_type":0,"onestop_id":"s-9q9p1wxf72-macarthur","stop_code":"","stop_desc":"","stop_id":"MCAR","stop_name":"MacArthur","stop_timezone":"","stop_url":"http://www.bart.gov/stations/MCAR/","wheelchair_boarding":1,"zone_id":"MCAR"}]}`,
			"",
			nil,
		},
		{
			// just ensure this query completes successfully; checking coordinates is a pain and flaky.
			"geometry",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {geometry} }`,
			vars,
			``,
			"stops.0.geometry.type",
			[]string{"Point"},
		},
		{
			"feed_version",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {feed_version_sha1} }`,
			vars,
			`{"stops":[{"feed_version_sha1":"e535eb2b3b9ac3ef15d82c56575e914575e732e0"}]}`,
			"",
			nil,
		},
		{
			"route_stops",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {route_stops{route{route_id route_short_name}}} }`,
			vars,
			``,
			"stops.0.route_stops.#.route.route_id",
			[]string{"01", "03", "07"},
		},
		{
			"where near 10m",
			`query {stops(where:{near:{lon:-122.407974,lat:37.784471,radius:10.0}}) {stop_id onestop_id geometry}}`,
			vars,
			``,
			"stops.#.stop_id",
			[]string{"POWL"},
		},
		{
			"where near 2000m",
			`query {stops(where:{near:{lon:-122.407974,lat:37.784471,radius:2000.0}}) {stop_id onestop_id geometry}}`,
			vars,
			``,
			"stops.#.stop_id",
			[]string{"70011", "70012", "CIVC", "EMBR", "MONT", "POWL"},
		},
		{
			"where within polygon",
			`query{stops(where:{within:{type:"Polygon",coordinates:[[[-122.396,37.8],[-122.408,37.79],[-122.393,37.778],[-122.38,37.787],[-122.396,37.8]]]}}){id stop_id}}`,
			hw{},
			``,
			"stops.#.stop_id",
			[]string{"EMBR", "MONT"},
		},
		{
			"where onestop_id",
			`query{stops(where:{onestop_id:"s-9q9k658fd1-sanjosediridoncaltrain"}) {stop_id} }`,
			vars,
			``,
			"stops.0.stop_id",
			[]string{"70262"},
		},
		{
			"where feed_version_sha1",
			`query($feed_version_sha1:String!) { stops(where:{feed_version_sha1:$feed_version_sha1}) { stop_id } }`, // just check BART
			hw{"feed_version_sha1": "e535eb2b3b9ac3ef15d82c56575e914575e732e0"},
			``,
			"stops.#.stop_id",
			bartStops,
		},
		{
			"where feed_onestop_id",
			`query{stops(where:{feed_onestop_id:"BA"}) { stop_id } }`, // just check BART
			hw{},
			``,
			"stops.#.stop_id",
			bartStops,
		},
		{
			"where stop_id",
			`query{stops(where:{stop_id:"12TH"}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			[]string{"12TH"},
		},
		{
			"where search",
			`query{stops(where:{search:"macarthur"}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			[]string{"MCAR", "MCAR_S"},
		},
		{
			"where search 2",
			`query{stops(where:{search:"ftvl"}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			[]string{"FTVL"},
		},
		{
			"where search 3",
			`query{stops(where:{search:"warm springs"}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			[]string{"WARM"},
		},
		// served_by_route_types
		// {
		// 	"served_by_route_types=0",
		// 	`query{stops(where:{served_by_route_types:[0]}) { stop_id } }`,
		// 	hw{},
		// 	``,
		// 	"stops.#.stop_id",
		// 	[]string{},
		// },
		// {
		// 	"served_by_route_types=1",
		// 	`query{stops(where:{served_by_route_types:[1]}) { stop_id } }`,
		// 	hw{},
		// 	``,
		// 	"stops.#.stop_id",
		// 	bartStops,
		// },
		// {
		// 	"served_by_route_types=2",
		// 	`query{stops(where:{served_by_route_types:[2]}) { stop_id } }`,
		// 	hw{},
		// 	``,
		// 	"stops.#.stop_id",
		// 	caltrainRailStops,
		// },
		// {
		// 	"served_by_route_types=3",
		// 	`query{stops(where:{served_by_route_types:[3]}) { stop_id } }`,
		// 	hw{},
		// 	``,
		// 	"stops.#.stop_id",
		// 	caltrainBusStops,
		// },
		// {
		// 	"served_by_route_types=0,1,2,3",
		// 	`query{stops(where:{served_by_route_types:[0,1,2,3]}) { stop_id } }`,
		// 	hw{},
		// 	``,
		// 	"stops.#.stop_id",
		// 	allStops,
		// },
		// {
		// 	"served_by_route_types=2,3",
		// 	`query{stops(where:{served_by_route_types:[2,3]}) { stop_id } }`,
		// 	hw{},
		// 	``,
		// 	"stops.#.stop_id",
		// 	caltrainStops,
		// },
		// served_by_onestop_ids
		{
			"served_by_onestop_ids=o-9q9-bayarearapidtransit",
			`query{stops(where:{served_by_onestop_ids:["o-9q9-bayarearapidtransit"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			bartStops,
		},
		{
			"served_by_onestop_ids=o-9q9-caltrain",
			`query{stops(where:{served_by_onestop_ids:["o-9q9-caltrain"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			// caltrain stops minus a couple non-service stops
			caltrainStops,
		},
		{
			"served_by_onestop_ids=r-9q9-antioch~sfia~millbrae",
			`query{stops(where:{served_by_onestop_ids:["r-9q9-antioch~sfia~millbrae"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			// yellow line stops
			[]string{"12TH", "16TH", "19TH", "19TH_N", "24TH", "ANTC", "BALB", "CIVC", "COLM", "CONC", "DALY", "EMBR", "GLEN", "LAFY", "MCAR", "MCAR_S", "MLBR", "MONT", "NCON", "ORIN", "PITT", "PCTR", "PHIL", "POWL", "ROCK", "SBRN", "SFIA", "SSAN", "WCRK", "WOAK"},
		},
		{
			"served_by_onestop_ids=r-9q9-antioch~sfia~millbrae,r-9q8y-richmond~dalycity~millbrae",
			`query{stops(where:{served_by_onestop_ids:["r-9q9-antioch~sfia~millbrae","r-9q8y-richmond~dalycity~millbrae"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			// combination of yellow and red line stops
			[]string{"12TH", "16TH", "19TH", "19TH_N", "24TH", "ANTC", "ASHB", "BALB", "CIVC", "COLM", "CONC", "DALY", "DBRK", "DELN", "PLZA", "EMBR", "GLEN", "LAFY", "MCAR", "MCAR_S", "MLBR", "MONT", "NBRK", "NCON", "ORIN", "PITT", "PCTR", "PHIL", "POWL", "RICH", "ROCK", "SBRN", "SFIA", "SSAN", "WCRK", "WOAK"},
		},
		{
			"served_by_onestop_ids=o-9q9-bayarearapidtransit,r-9q9-antioch~sfia~millbrae",
			`query{stops(where:{served_by_onestop_ids:["o-9q9-bayarearapidtransit","r-9q9-antioch~sfia~millbrae"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			// all bart stops
			bartStops,
		},
		{
			"served_by_onestop_ids=o-9q9-bayarearapidtransit,o-9q9-caltrain",
			`query{stops(where:{served_by_onestop_ids:["o-9q9-bayarearapidtransit","o-9q9-caltrain"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			// all stops
			allStops,
		},
		// {
		// 	"served_by_route_types=2,served_by_onestop_ids=o-9q9-bayarearapidtransit,o-9q9-caltrain",
		// 	`query{stops(where:{served_by_onestop_ids:["o-9q9-bayarearapidtransit","o-9q9-caltrain"], served_by_route_types:[2]}) { stop_id } }`,
		// 	hw{},
		// 	``,
		// 	"stops.#.stop_id",
		// 	caltrainRailStops,
		// },
		// TODO: parent, children; test data has no stations.
		// TODO: level, pathways_from_stop, pathways_to_stop: test data has no pathways...
		// TODO: census_geographies
		// stop_times
		{
			"stop_times",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {stop_times { trip { trip_id} }} }`,
			hw{"stop_id": "70302"}, // Morgan hill
			``,
			"stops.0.stop_times.#.trip.trip_id",
			[]string{"268", "274", "156"},
		},
		{
			"stop_times where weekday_morning",
			`query($stop_id: String!, $service_date:Date!) {  stops(where:{stop_id:$stop_id}) {stop_times(where:{service_date:$service_date, start_time:21600, end_time:25200}) { trip { trip_id} }} }`,
			hw{"stop_id": "MCAR", "service_date": "2018-05-29"},
			``,
			"stops.0.stop_times.#.trip.trip_id",
			[]string{"3830503WKDY", "3850526WKDY", "3610541WKDY", "3630556WKDY", "3650611WKDY", "2210533WKDY", "2230548WKDY", "2250603WKDY", "2270618WKDY", "4410518WKDY", "4430533WKDY", "4450548WKDY", "4470603WKDY"},
		},
		{
			"stop_times where sunday_morning",
			`query($stop_id: String!, $service_date:Date!) {  stops(where:{stop_id:$stop_id}) {stop_times(where:{service_date:$service_date, start_time:21600, end_time:36000}) { trip { trip_id} }} }`,
			hw{"stop_id": "MCAR", "service_date": "2018-05-27"},
			``,
			"stops.0.stop_times.#.trip.trip_id",
			[]string{"3730756SUN", "3750757SUN", "3770801SUN", "3790821SUN", "3610841SUN", "3630901SUN", "2230800SUN", "2250748SUN", "2270808SUN", "2290828SUN", "2310848SUN", "2330908SUN"},
		},
		{
			"stop_times where saturday_evening",
			`query($stop_id: String!, $service_date:Date!) {  stops(where:{stop_id:$stop_id}) {stop_times(where:{service_date:$service_date, start_time:57600, end_time:72000}) { trip { trip_id} }} }`,
			hw{"stop_id": "MCAR", "service_date": "2018-05-26"},
			``,
			"stops.0.stop_times.#.trip.trip_id",
			[]string{"3611521SAT", "3631541SAT", "3651601SAT", "3671621SAT", "3691641SAT", "3711701SAT", "3731721SAT", "3751741SAT", "3771801SAT", "3791821SAT", "3611841SAT", "3631901SAT", "2231528SAT", "2251548SAT", "2271608SAT", "2291628SAT", "2311648SAT", "2331708SAT", "2351728SAT", "2211748SAT", "2231808SAT", "2251828SAT", "2271848SAT", "2291908SAT", "4471533SAT", "4491553SAT", "4511613SAT", "4531633SAT", "4411653SAT", "4431713SAT", "4451733SAT", "4471753SAT", "4491813SAT", "4511833SAT", "4531853SAT"},
		},
		// TODO: census_geographies
		// TODO: route_stop_buffer
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}
