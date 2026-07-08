# All binaries owned by a given user
data "cumulocity_binaries" "mine" {
  owner = "admin"
}

output "binary_names" {
  value = [for b in data.cumulocity_binaries.mine.binaries : b.name]
}

output "total_binary_bytes" {
  value = sum([for b in data.cumulocity_binaries.mine.binaries : b.length])
}
