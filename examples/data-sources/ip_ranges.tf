provider "fastly" {
  no_auth = true
}

data "fastly_ip_ranges" "fastly" {}

resource "aws_security_group" "from_fastly" {
  name = "from_fastly"

  ingress {
    from_port         = "443"
    to_port           = "443"
    protocol          = "tcp"
    cidr_blocks       = data.fastly_ip_ranges.fastly.cidr_blocks
    ipv6_cidr_blocks  = data.fastly_ip_ranges.fastly.ipv6_cidr_blocks
  }
}