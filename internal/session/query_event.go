package session

func queryEventPayload(sid string, ev QueryRunEvent) map[string]any {
	payload := map[string]any{
		"sid": sid,
		"qid": ev.QID,
	}
	if ev.Done {
		payload["done"] = true
	}
	if ev.Cancelled {
		payload["cancelled"] = true
	}
	if ev.StmtIndex > 0 {
		payload["stmt_index"] = ev.StmtIndex
	}
	if ev.StmtTotal > 0 {
		payload["stmt_total"] = ev.StmtTotal
	}
	if ev.StmtStart > 0 {
		payload["stmt_start"] = ev.StmtStart
	}
	if ev.RID != "" {
		payload["rid"] = ev.RID
	}
	if ev.RowsAffected > 0 {
		payload["rows_affected"] = ev.RowsAffected
	}
	if ev.DurationMs > 0 {
		payload["duration_ms"] = ev.DurationMs
	}
	if ev.Error != "" {
		payload["error"] = ev.Error
		payload["error_detail"] = ev.ErrorDetail
		payload["position"] = ev.Position
	}
	if ev.NeedsConfirm {
		payload["needs_confirm"] = true
		payload["position"] = ev.Position
	}
	return payload
}
