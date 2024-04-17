package model

type ConfigShard struct {
	Name     string `bson:"_id"`
	Draining bool   `bson:"draining"`
}

type ConfigBalancerSettings struct {
	Stopped bool `bson:"stopped"`
}

type ConfigDatabase struct {
	Name        string `bson:"_id"`
	Partitioned bool   `bson:"partitioned"`
}

type ConfigChunks []ConfigChunk

type ConfigChunk struct {
	Ns      string `bson:"-"`
	Shard   string `bson:"_id"`
	NChunks int    `bson:"nChunks"`
}

type ConfigChunkMoves struct {
	Ns      string `bson:"_id"`
	NChunks int    `bson:"nChunks"`
}
