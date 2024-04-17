package model

type ListDatabasesDoc struct {
	Databases []struct {
		Name string `bson:"name"`
	} `bson:"databases"`
}

type ListCollectionsDoc struct {
	Cursor struct {
		FirstBatch []Collection `bson:"firstBatch"`
	} `bson:"cursor"`
}

type Collection struct {
	Name string `bson:"name"`
	Type string `bson:"type"`
	Info struct {
		ReadOnly bool        `bson:"readOnly"`
		UUID     interface{} `bson:"uuid"`
	} `bson:"info"`
}
