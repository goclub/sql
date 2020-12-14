package sq

type Model interface {
	TableName() string
	BeforeCreate()
}

type Relation interface {
	FormTable () string
	RelationJoin () []Join
}