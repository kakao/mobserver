package model

type MongoReplRoleType int32

const (
	REPL_STARTUP    MongoReplRoleType = 0  // Not yet an active member of any set. All members start up in this state. The mongod parses the replica set configuration document while in STARTUP.
	REPL_PRIMARY    MongoReplRoleType = 1  // The member in state primary is the only member that can accept write operations.
	REPL_SECONDARY  MongoReplRoleType = 2  // A member in state secondary is replicating the data store. Data is available for reads, although they may be stale.
	REPL_RECOVERING MongoReplRoleType = 3  // Can vote. Members either perform startup self-checks, or transition from completing a rollback or resync.
	REPL_STARTUP2   MongoReplRoleType = 5  // The member has joined the set and is running an initial sync.
	REPL_UNKNOWN    MongoReplRoleType = 6  // The member’s state, as seen from another member of the set, is not yet known.
	REPL_ARBITER    MongoReplRoleType = 7  // Arbiters do not replicate data and exist solely to participate in elections.
	REPL_DOWN       MongoReplRoleType = 8  // The member, as seen from another member of the set, is unreachable.
	REPL_ROLLBACK   MongoReplRoleType = 9  // This member is actively performing a rollback. Data is not available for reads.
	REPL_REMOVED    MongoReplRoleType = 10 // This member was once in a replica set but was subsequently removed.
)

func (r MongoReplRoleType) IsOddState() bool {
	switch r {
	case REPL_PRIMARY, REPL_SECONDARY, REPL_ARBITER, REPL_STARTUP2:
		return false
	default:
		return true
	}
}

func (r MongoReplRoleType) String() string {
	switch r {
	case REPL_STARTUP:
		return "STARTUP"
	case REPL_PRIMARY:
		return "PRIMARY"
	case REPL_SECONDARY:
		return "SECONDARY"
	case REPL_RECOVERING:
		return "RECOVERING"
	case REPL_STARTUP2:
		return "STARTUP2"
	case REPL_UNKNOWN:
		return "UNKNOWN"
	case REPL_ARBITER:
		return "ARBITER"
	case REPL_DOWN:
		return "DOWN"
	case REPL_ROLLBACK:
		return "ROLLBACK"
	case REPL_REMOVED:
		return "REMOVED"
	default:
		return "UNKNOWN"
	}
}
