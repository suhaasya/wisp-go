package pager

// Cursor-based query paging for PostgreSQL DECLARE/FETCH and MySQL streaming
// is implemented via queryFetcher in query.go (forward-only rows with re-query on backward seek).
