package dejavu

import "fmt"

const (
	HistoryTableName       = "deja_vu_history"
	HistoryColumnName      = "name"
	HistoryColumnStartedAt = "started_at"
	HistoryColumnDuration  = "duration_ms"
	HistoryColumnChecksum  = "checksum"
)

const (
	LockTableName      = "deja_vu_lock"
	LockColumnID       = "id"
	LockColumnHostname = "hostname"
	LockColumnPid      = "pid"
	LockColumnSince    = "since"
)

type Statements interface {
	fmt.Stringer

	CountFromTable(name string) *Statement

	CreateHistoryTable() *Statement
	CreateLockTable() *Statement

	Lock(lck Lock) *Statement
	Unlock(lck Lock) *Statement

	History() *Statement
	Log(mig Migration) *Statement
}

type DefaultStatements struct{}

func (s DefaultStatements) CountFromTable(name string) *Statement {
	return NewStatement("select count(1) from " + name)
}

func (s DefaultStatements) CreateHistoryTable() *Statement {
	return NewStatement(
		`create table %s (
			%s varchar(512) not null,
			%s timestamp    not null,
			%s int          not null,
			%s char(43)     not null,
			constraint %s_pk primary key (name)
		)`,
		HistoryTableName,
		HistoryColumnName,
		HistoryColumnStartedAt,
		HistoryColumnDuration,
		HistoryColumnChecksum,
		HistoryTableName,
	)
}

func (s DefaultStatements) CreateLockTable() *Statement {
	return NewStatement(
		`create table %s (
			%s int          not null,
			%s varchar(128) not null,
			%s int          not null,
			%s timestamp    not null,
			constraint %s_pk primary key (id)
		)`,
		LockTableName,
		LockColumnID,
		LockColumnHostname,
		LockColumnPid,
		LockColumnSince,
		LockTableName,
	)
}

func (s DefaultStatements) Lock(lck Lock) *Statement {
	return NewStatement(
		"insert into %s (%s, %s, %s, %s) values (:id, :hostname, :pid, :since)",
		LockTableName,
		LockColumnID,
		LockColumnHostname,
		LockColumnPid,
		LockColumnSince,
	).
		Arg("id", lck.id).
		Arg("hostname", lck.hostname).
		Arg("pid", lck.pid).
		Arg("since", lck.since)
}

func (s DefaultStatements) Unlock(lck Lock) *Statement {
	return NewStatement(
		"delete from %s where %s = :id",
		LockTableName,
		LockColumnID,
	).
		Arg("id", lck.id)
}

func (s DefaultStatements) History() *Statement {
	return NewStatement(
		"select %s, %s, %s, %s from %s order by %s",
		HistoryColumnName,
		HistoryColumnStartedAt,
		HistoryColumnDuration,
		HistoryColumnChecksum,
		HistoryTableName,
		HistoryColumnName,
	)
}

func (s DefaultStatements) Log(mig Migration) *Statement {
	return NewStatement(
		"insert into %s (%s, %s, %s, %s) values (:name, :start, :duration_ms, :checksum)",
		HistoryTableName,
		HistoryColumnName,
		HistoryColumnStartedAt,
		HistoryColumnDuration,
		HistoryColumnChecksum,
	).
		Arg("name", mig.Name).
		Arg("start", mig.Start).
		Arg("duration_ms", mig.DurationMs).
		Arg("checksum", mig.Checksum)
}

func (s DefaultStatements) String() string {
	return "Default SQL statements"
}
