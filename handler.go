package mongolog

import (
	"context"
	"log/slog"
	"maps"
	"math"
	"slices"
)

type Inserter interface {
	Insert(ctx context.Context, collection string, data map[string]interface{}) error
}

// MongoHandler implement slog.Handler interface
// and logs to mongodb
type MongoHandler struct {
	db                Inserter
	collection        string
	Level             slog.Level
	groups            []string
	preformattedAttrs map[string]interface{}
}

// NewMongoHandler create new MongoHandler
// with default mongo collection name "app"
func NewMongoHandler(db Inserter, collection string, level slog.Level) *MongoHandler {
	if collection == "" {
		collection = "app"
	}
	return &MongoHandler{
		Level:             level,
		db:                db,
		collection:        collection,
		preformattedAttrs: make(map[string]interface{}),
	}
}

func (h *MongoHandler) Handle(ctx context.Context, r slog.Record) error {
	// data is finish place for all log data
	data := make(map[string]interface{})

	// convert slog.Record to header map (Time, Level, Msg) and attrs map (Attrs)
	header, attrs := recordToMap(r)

	// Copy headers to data
	maps.Copy(data, header)

	// Clone preformatted attrs to don't change original
	copyAttrs := maps.Clone(h.preformattedAttrs)

	// Copy attrs to last group of preformatted attrs
	maps.Copy(lastGroup(h.groups, copyAttrs), attrs)

	// Copy attrs to data
	maps.Copy(data, copyAttrs)

	// Insert this map to mongo
	if err := h.db.Insert(ctx, h.collection, data); err != nil {
		ErrNotify(err)
	}
	return nil
}

func (h *MongoHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.Level <= level
}

func (h *MongoHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// clone original handler
	clone := h.clone()

	// get last group of preformatted attrs
	lg := lastGroup(clone.groups, clone.preformattedAttrs)

	// Convert types and add to last group
	for _, a := range attrs {
		// switch case that check kind of a.Value and convert him.
		switch a.Value.Kind() {
		case slog.KindString:
			lg[a.Key] = a.Value.String()
		case slog.KindInt64:
			// int64 in mongodb stores like Long("12345")
			// int32 just like a number
			if v := a.Value.Int64(); v > math.MaxInt32 || v < -math.MaxInt32 {
				lg[a.Key] = v
			} else {
				lg[a.Key] = int(v)
			}
		case slog.KindFloat64:
			lg[a.Key] = a.Value.Float64()
		case slog.KindBool:
			lg[a.Key] = a.Value.Bool()
		case slog.KindTime:
			lg[a.Key] = a.Value.Time()
		case slog.KindAny:
			lg[a.Key] = a.Value.Any()
		}
	}
	return clone
}

func (h *MongoHandler) WithGroup(name string) slog.Handler {
	// clone original handler
	clone := h.clone()
	// add new group
	lastGroup(clone.groups, clone.preformattedAttrs)[name] = make(map[string]interface{})
	// add group name to groups slice
	clone.groups = append(clone.groups, name)
	return clone
}

func (h *MongoHandler) clone() *MongoHandler {
	// create new MongoHandler from h
	return &MongoHandler{
		db:                h.db,
		collection:        h.collection,
		Level:             h.Level,
		groups:            slices.Clip(h.groups),
		preformattedAttrs: maps.Clone(h.preformattedAttrs),
	}
}

// recordToMap convert slog.Record to map[string]interface{}
func recordToMap(r slog.Record) (header map[string]interface{}, attrs map[string]interface{}) {
	// Create header
	header = make(map[string]interface{})
	header["time"] = r.Time
	header["msg"] = r.Message
	header["level"] = r.Level

	// Create attrs
	//
	// OH NO, IT'S NOT DRY
	// CONVERTING LOGIC REPEATS AT LINE 74
	// CRINGE CRINGE CRINGE
	//
	// Fuck yourself.
	// I don't want make a million adapters to observe DRY
	// cuz it will make my code more complicated.
	// Copypaste suprimacy.
	attrs = make(map[string]interface{})
	r.Attrs(func(a slog.Attr) bool {
		// switch case that check kind of a.Value and convert him
		switch a.Value.Kind() {
		case slog.KindString:
			attrs[a.Key] = a.Value.String()
		case slog.KindInt64:
			if v := a.Value.Int64(); v > int64(math.Pow(2, 32)) {
				attrs[a.Key] = v
			} else {
				attrs[a.Key] = int(v)
			}
		case slog.KindFloat64:
			attrs[a.Key] = a.Value.Float64()
		case slog.KindBool:
			attrs[a.Key] = a.Value.Bool()
		case slog.KindTime:
			attrs[a.Key] = a.Value.Time()
		case slog.KindAny:
			attrs[a.Key] = a.Value.Any()
		}
		return true
	})
	return header, attrs
}

// lastGroup get last group of groups
func lastGroup(groups []string, group map[string]interface{}) map[string]interface{} {
	if len(groups) == 0 {
		return group
	}
	for _, name := range groups {
		var ok bool
		group, ok = group[name].(map[string]interface{})
		if !ok {
			ErrNotify("group " + name + " is not map[string]interface{}")
			break
		}
	}
	return group
}
