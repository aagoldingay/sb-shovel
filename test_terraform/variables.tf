locals {
  tags = {
    product = "sb-shovel"
    source  = "terraform"
  }
}

variable "rg_name" {
  type = string
}

variable "rg_location" {
  type = string
}

variable "sb_name" {
  type = string
}

variable "q_name" {
  type = string
}
