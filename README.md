# mongolog

## Install
    go get github.com/Richtermnd/mongolog

## Example
```golang
func main() {
    db, err := mongolog.NewMongoDB(mongolog.Config{
        URI:    "mongodb://root:password@localhost:27017",
        DBName: "logs",
    })
    if err != nil {
        log.Fatal(err)
    }

    // For different apps you can use different mongodb collections.
    handler := mongolog.NewMongoHandler(inserter, "app", slog.LevelDebug)
    log := slog.New(handler)
    log.Info("test1", slog.String("key", "value"))

    log = log.With(slog.String("attr", "value"))
    log.Info("test2", slog.String("key", "value"))

    log = log.WithGroup("group1").With(slog.String("attr1", "value1")).WithGroup("group2").With(slog.String("attr2", "value2"))
    log.Info("test3", slog.String("key", "value"))
}
```