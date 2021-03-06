package main

import (
    "fmt"
    "log"
    "net/http"
    "time"

    "vchaind/handler"
    "github.com/gorilla/mux"
)

type Route struct {
    Method      string
    Pattern     string
    HandlerFunc http.HandlerFunc
}

var routes = []Route{
    Route{"GET",  "/api/v1/ping", handler.Ping    },
    Route{"POST", "/api/v1/data", handler.PostData},

    Route{"GET",  "/api/v1/services",                           handler.GetServices                },
    Route{"GET",  "/api/v1/services-chain",                     handler.GetServiceChain            },
    Route{"GET",  "/api/v1/services/{name}/request-categories", handler.GetServiceRequestCategories},

    Route{"GET",  "/api/v1/invoke-chains",                                         handler.GetAllInvokeChains        },
    Route{"GET",  "/api/v1/invoke-chains/{service}/{category}",                    handler.GetInvokeChains           },
    Route{"GET",  "/api/v1/invoke-chains/{service}/{category}/{id}",               handler.GetInvokeChain            },
    Route{"GET",  "/api/v1/invoke-chains/{service}/{category}/{id}/root-requests", handler.GetInvokeChainRootRequests},

    Route{"GET",  "/api/v1/request-overview",              handler.GetRequestOverview    },
    Route{"GET",  "/api/v1/request-types",                 handler.GetRequestTypes       },
    Route{"GET",  "/api/v1/requests",                      handler.GetRequests           },
    Route{"GET",  "/api/v1/requests/{uuid}",               handler.GetRequest            },
    Route{"GET",  "/api/v1/requests/{uuid}/invoke-chain",  handler.GetRequestInvokeChain },
    Route{"GET",  "/api/v1/requests/{uuid}/request-group", handler.GetRequestRequestGroup},
    Route{"GET",  "/api/v1/requests/{uuid}/root-request",  handler.GetRequestRootRequest },
    Route{"GET",  "/api/v1/requests/{uuid}/parent",        handler.GetRequestParent      },
    Route{"GET",  "/api/v1/requests/{uuid}/children",      handler.GetRequestChildren    },
}

type InnerResponseWriter struct {
    StatusCode int
    isSet      bool
    http.ResponseWriter
}

func (i *InnerResponseWriter) WriteHeader(status int) {
    if !i.isSet {
        i.StatusCode = status
        i.isSet = true
    }

    i.ResponseWriter.WriteHeader(status)
}

func (i *InnerResponseWriter) Write(b []byte) (int, error) {
    i.isSet = true
    return i.ResponseWriter.Write(b)
}

func wrapper(inner http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        s := time.Now()
        wr := &InnerResponseWriter{
            StatusCode:     200,
            isSet:          false,
            ResponseWriter: w,
        }

        defer func() {
            if err := recover(); err != nil {
                wr.WriteHeader(http.StatusInternalServerError)
                log.Printf("Panic: %v\n", err)
                fmt.Fprintf(w, fmt.Sprintln(err))
            }

            d := time.Now().Sub(s)
            log.Printf("%s %s %d %s\n", r.Method, r.RequestURI, wr.StatusCode, d.String())
        }()

        inner.ServeHTTP(wr, r)
    })
}

func NewRouter() *mux.Router {
    router := mux.NewRouter()
    for _, route := range routes {
        router.Methods(route.Method).Path(route.Pattern).HandlerFunc(wrapper(route.HandlerFunc))
    }

    return router
}
