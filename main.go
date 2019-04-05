package main

import (
	"net/http"
	"os"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	httptransport "github.com/go-kit/kit/transport/http"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)

	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)
	countResult := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "count_result",
		Help:      "The result of each count method.",
	}, []string{}) // no fields here

	var svc StringService
	svc = stringService{}
	bookSvc := bookService{}
	hostSvc := hostService{}
	svc = loggingMiddleware{logger, svc}
	svc = instrumentingMiddleware{requestCount, requestLatency, countResult, svc}

	uppercaseHandler := httptransport.NewServer(
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
		encodeResponse,
	)

	countHandler := httptransport.NewServer(
		makeCountEndpoint(svc),
		decodeCountRequest,
		encodeResponse,
	)

	//-----------------------------------
	booksHandler := httptransport.NewServer(
		makeBooksEndpoint(bookSvc),
		decodeBooksRequest,
		encodeResponse,
	)

	bookHandler := httptransport.NewServer(
		makeBookEndpoint(bookSvc),
		decodeBookRequest,
		encodeResponse,
	)

	setbookHandler := httptransport.NewServer(
		makeSetBookEndpoint(bookSvc),
		decodeSetBookRequest,
		encodeResponse,
	)
	//--------------------------------
	hostHandler := httptransport.NewServer(
		makeHostEndpoint(hostSvc),
		decodeHostRequest,
		encodeResponse,
	)

	http.Handle("/uppercase", uppercaseHandler)
	http.Handle("/count", countHandler)
	http.Handle("/metrics", promhttp.Handler())

	//--------------------------------
	http.Handle("/books", booksHandler)
	http.Handle("/book", bookHandler)
	http.Handle("/setbook", setbookHandler)

	// ---------------------------------------
	http.Handle("/hostname", hostHandler)

	logger.Log("msg", "HTTP", "addr", ":5050")
	logger.Log("err", http.ListenAndServe(":5050", nil))
}
