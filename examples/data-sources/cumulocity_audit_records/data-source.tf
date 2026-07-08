# Audit records for actions carried out by a specific user
data "cumulocity_audit_records" "by_user" {
  user = "admin"
  type = "Operation"
}

output "audit_activities" {
  value = [for r in data.cumulocity_audit_records.by_user.audit_records : r.activity]
}
