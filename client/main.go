package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/bsm/redislock"
	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sigmavirus24/circuitry"
	redisbackend "github.com/sigmavirus24/circuitry/backends/redis"
	"github.com/sigmavirus24/circuitry/log"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"golang.org/x/net/context"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var (
	logger       = log.NewLogrus(logrus.New())
	client       = resty.New()
	breaker      circuitry.CircuitBreaker
	goBreaker    *gobreaker.CircuitBreaker
	urlServer    = "http://localhost:8081/data"
	totalRequest = 0
)

func main() {
	initGoBreaker()

	//setup hystrix
	hystrix.ConfigureCommand("api_get_hystrix", hystrix.CommandConfig{
		Timeout:                5000, // 5s timeout func
		MaxConcurrentRequests:  20,
		ErrorPercentThreshold:  50,
		SleepWindow:            10000, // 10s to sleep after tripped to open
		RequestVolumeThreshold: 20,
	})

	http.HandleFunc("/circuitry", handlerWithCircuitry)

	http.HandleFunc("/gobreaker", handlerWithGoBreaker)

	http.HandleFunc("/hystrix", handlerHystrix)

	//Hystrix Dashboard Metrics
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go http.ListenAndServe(net.JoinHostPort("", "81"), hystrixStreamHandler)

	// Start the HTTP server on port 8080
	fmt.Println("Server listening on port 8080")
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func handlerHystrix(w http.ResponseWriter, r *http.Request) {
	totalRequest++
	defer logger.Info(fmt.Sprintf("TOTAL REQUEST : %d", totalRequest))
	data := Response{}

	err := hystrix.DoC(r.Context(), "api_get_hystrix", func(c context.Context) error {
		// talk to other services
		logger.Info("EXECUTE HTTP GET BY HYSTRIX")
		response, err := client.R().Get(urlServer)
		if err != nil {
			return err
		}

		if response.StatusCode() >= http.StatusInternalServerError {
			fmt.Println("RESPONSE ERROR", string(response.Body()))
			return errors.New("internal server error")
		}

		fmt.Println("RESPONSE SUCCESS", string(response.Body()))
		return nil
	}, func(c context.Context, srcErr error) error {
		logger.Info("EXECUTE FALLBACK HYSTRIX")
		logger.Info(fmt.Sprintf("Source Error : %s", srcErr.Error()))
		return srcErr
	})
	if err != nil {
		data.Message = err.Error()
		responseJson(w, data)
		return
	}

	data.Success = true
	data.Message = "success"
	responseJson(w, data)
}

func handlerWithCircuitry(w http.ResponseWriter, r *http.Request) {
	// Setup Circuitry
	factory, err := CreateCircuitBreakerFactory(logger)
	if err != nil {
		fmt.Println("Error factory settings", err)
		return
	}
	breaker = factory.BreakerFor("api_get_circuitry", map[string]any{})

	data := Response{}

	_, workErr, err := breaker.Execute(r.Context(), func() (any, error) {
		logger.Info("EXECUTE HTTP GET BY CIRCUITRY")
		response, err := client.SetTimeout(5 * time.Second).R().Get(urlServer)
		if err != nil {
			return nil, err
		}

		if response.StatusCode() >= http.StatusInternalServerError {
			return nil, errors.New("internal server error")
		}

		fmt.Println("RESPONSE SUCCESS", string(response.Body()))

		return nil, nil
	})

	if err != nil {
		if errors.Is(err, circuitry.ErrCircuitBreakerOpen) {
			logger.Error("CIRCUIT BREAKER IS OPEN")
		} else {
			logger.Error(fmt.Sprintf("ERROR EXECUTE => %s", err.Error()))
		}
		data.Message = err.Error()
		responseJson(w, data)
		return
	}
	if workErr != nil {
		logger.Error(fmt.Sprintf("ERROR WORKER => %s", workErr.Error()))
		data.Message = workErr.Error()
		responseJson(w, data)
		return
	}

	data.Success = true
	data.Message = "success"
	responseJson(w, data)
}

func handlerWithGoBreaker(w http.ResponseWriter, r *http.Request) {
	data := Response{}

	_, circuitBreakerErr := goBreaker.Execute(func() (interface{}, error) {
		logger.Info("EXECUTE HTTP GET GOBREAKER")
		response, err := client.SetTimeout(5 * time.Second).R().Get(urlServer)
		if err != nil {
			return nil, err
		}

		if response.StatusCode() >= http.StatusInternalServerError {
			return nil, errors.New("internal server error")
		}

		fmt.Println("RESPONSE SUCCESS", string(response.Body()))
		return nil, err
	})

	if circuitBreakerErr != nil {
		data.Message = circuitBreakerErr.Error()
		responseJson(w, data)
		return
	}

	data.Success = true
	data.Message = "success"
	responseJson(w, data)
}

func CreateCircuitBreakerFactory(logger log.Logger) (*circuitry.CircuitBreakerFactory, error) {
	settings, err := circuitry.NewFactorySettings(
		// circuitry.WithStorageBackend(backend),
		// backends.WithInMemoryBackend(),
		redisbackend.WithRedisBackend(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}, &redislock.Options{
			RetryStrategy: redislock.NoRetry(),
		}, 30*time.Second),
		circuitry.WithDefaultNameFunc(),
		circuitry.WithDefaultTripFunc(),
		circuitry.WithDefaultFallbackErrorMatcher(),
		circuitry.WithFailureCountThreshold(5),
		circuitry.WithCloseThreshold(5),
		circuitry.WithAllowAfter(1*time.Minute),
		// circuitry.WithCyclicClearAfter(12*time.Hour),
		circuitry.WithStateChangeCallback(func(name string, circuitContext map[string]any, from, to circuitry.CircuitState) {
			logger.
				WithFields(circuitContext). // Ensure no sensitive information is logged here
				WithFields(log.Fields{
					"name": name,
					"from": from.String(),
					"to":   to.String(),
				}).Debug("state transition")
		}),
	)
	if err != nil {
		return nil, err
	}
	return circuitry.NewCircuitBreakerFactory(settings), nil
}

func responseJson(w http.ResponseWriter, data interface{}) {
	// Encode the data as JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func initGoBreaker() {
	// Setup Gobreaker
	cbSettings := gobreaker.Settings{
		Name:        "api_get_gobreaker",
		MaxRequests: 10,
		Interval:    1 * time.Minute,
		Timeout:     1 * time.Minute,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 4
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Info(fmt.Sprintf("BREAKER NAME : %s", name))
			logger.Info(fmt.Sprintf("FROM STATE : %s", from.String()))
			logger.Info(fmt.Sprintf("TO STATE : %s", to.String()))
		},
	}
	goBreaker = gobreaker.NewCircuitBreaker(cbSettings)
}
