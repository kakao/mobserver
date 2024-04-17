package model

type TopField struct {
	Time  int64 `bson:"time"`
	Count int64 `bson:"count"`
}

type Top struct {
	Totals map[string]map[string]TopField `bson:"totals"`
}
