package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var serviceName = "ServiceATrace"

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func ginrun() {
	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)

	router.Run("localhost:8080")
}

func waitFunc(ctx context.Context) context.Context {
	ctx, span := otel.Tracer(serviceName).Start(ctx, "waitFunc")
	defer span.End()

	time.Sleep(1 * time.Second)
	return ctx
}

func errFunc(ctx context.Context) context.Context {
	ctx, span := otel.Tracer(serviceName).Start(ctx, "errFunc")
	defer span.End()

	bg := baggage.FromContext(ctx)
	fmt.Println(bg.Len())

	err := errors.New("this is temp error")
	span.RecordError(err)
	span.SetStatus(codes.Error, "Something not found")

	time.Sleep(1 * time.Second)
	return ctx
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	ctx, span := otel.Tracer(serviceName).Start(c.Request.Context(), "getAlbums")
	defer span.End()

	ctx = waitFunc(ctx)
	errFunc(ctx)

	c.IndentedJSON(http.StatusOK, albums)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	_, span := otel.Tracer(serviceName).Start(c.Request.Context(), "getAlbumByID", trace.WithAttributes(
		attribute.String("key", "value"),
		attribute.Bool("case", true),
	))
	defer span.End()

	id := c.Param("id")
	span.SetAttributes(attribute.String("id", id))

	span.AddEvent("Event-AsliConf",
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(attribute.String("AsliConf", "AsliHaiAsliHai")))

	// Loop through the list of albums, looking for
	// an album whose ID value matches the parameter.
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}
