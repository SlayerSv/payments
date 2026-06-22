# ==============================================================================
# 1. НАСТРОЙКА ПРОВАЙДЕРА CLOUD.RU EVOLUTION
# ==============================================================================
terraform {
  required_providers {
    cloudru = {
      source  = "cloud.ru/cloudru/cloud" # Официальный провайдер Evolution
      version = ">= 2.0.0"
    }
  }
}

provider "cloudru" {
  project_id  = var.project_id
  auth_key_id = var.auth_key_id  # Твой Key ID (логин сервисного аккаунта)
  auth_secret = var.auth_secret  # Твой Key Secret (пароль сервисного аккаунта)
}

# ==============================================================================
# 2. ПЕРЕМЕННЫЕ
# ==============================================================================
variable "project_id" {
  type        = string
  description = "Идентификатор проекта в Cloud.ru Evolution"
}

variable "auth_key_id" {
  type        = string
  description = "Key ID сервисного аккаунта"
}

variable "auth_secret" {
  type        = string
  sensitive   = true
  description = "Key Secret сервисного аккаунта"
}

variable "existing_subnet_id" {
  type        = string
  default     = "твоя-подсеть-id" # Скопируй ID подсети из консоли
  description = "ID существующей подсети в VPC"
}

variable "existing_postgres_volume_id" {
  type        = string
  default     = "твой-диск-id" # Скопируй ID предсозданного диска EVS
  description = "ID твоего существующего диска для базы данных"
}

# ==============================================================================
# 3. СЕТЕВЫЕ ИНТЕРФЕЙСЫ (В Evolution интерфейсы создаются отдельно)
# ==============================================================================

# Сетевой интерфейс для Master (VM-1)
resource "cloudru_evolution_compute_interface" "master_interface" {
  project_id = var.project_id
  subnet_id  = var.existing_subnet_id
}

# Сетевой интерфейс для Worker (VM-2)
resource "cloudru_evolution_compute_interface" "worker_interface" {
  project_id = var.project_id
  subnet_id  = var.existing_subnet_id
}

# Сетевой интерфейс для Data Node (VM-3)
resource "cloudru_evolution_compute_interface" "data_interface" {
  project_id = var.project_id
  subnet_id  = var.existing_subnet_id
}

# Сетевой интерфейс для Ops (VM-4)
resource "cloudru_evolution_compute_interface" "ops_interface" {
  project_id = var.project_id
  subnet_id  = var.existing_subnet_id
}

# ==============================================================================
# 4. ВИРТУАЛЬНЫЕ МАШИНЫ (В Evolution это ресурс cloudru_evolution_compute_vm)
# ==============================================================================

# VM-1: Master Node (2 vCPU, 4GB RAM) -> Сбалансированная
resource "cloudru_evolution_compute_vm" "vm_master" {
  project_id = var.project_id
  name       = "vm-1-master"

  zone_identifier = {
    name = "ru.AZ-2" # Твоя зона доступности из консоли
  }

  flavor_identifier = {
    name = "gen-2-4" # 2 ядра, 4 ГБ оперативной памяти
  }

  image = {
    id = "ubuntu-22-04-id" # ID образа Ubuntu 22.04 из твоего каталога
  }

  network_interfaces = [
    {
      interface_id = cloudru_evolution_compute_interface.master_interface.id
    }
  ]
}

# VM-2: Worker Node (2 vCPU, 8GB RAM) -> Сбалансированная (как твоя тестовая)
resource "cloudru_evolution_compute_vm" "vm_worker" {
  project_id = var.project_id
  name       = "vm-2-worker"

  zone_identifier = {
    name = "ru.AZ-2"
  }

  flavor_identifier = {
    name = "gen-2-8" # 2 ядра, 8 ГБ оперативной памяти (тот самый флейвор!)
  }

  image = {
    id = "ubuntu-22-04-id"
  }

  network_interfaces = [
    {
      interface_id = cloudru_evolution_compute_interface.worker_interface.id
    }
  ]
}

# VM-3: Data Node (2 vCPU, 8GB RAM + ПОДКЛЮЧЕННЫЙ ОБЛАЧНЫЙ ДИСК)
resource "cloudru_evolution_compute_vm" "vm_data" {
  project_id = var.project_id
  name       = "vm-3-data"

  zone_identifier = {
    name = "ru.AZ-2"
  }

  flavor_identifier = {
    name = "gen-2-8"
  }

  image = {
    id = "ubuntu-22-04-id"
  }

  network_interfaces = [
    {
      interface_id = cloudru_evolution_compute_interface.data_interface.id
    }
  ]

  # В Evolution диски монтируются прямо внутри ресурса VM! Очень удобно:
  disks = [
    {
      disk_id = var.existing_postgres_volume_id
    }
  ]
}

# VM-4: Ops Node (2 vCPU, 8GB RAM)
resource "cloudru_evolution_compute_vm" "vm_ops" {
  project_id = var.project_id
  name       = "vm-4-ops"

  zone_identifier = {
    name = "ru.AZ-2"
  }

  flavor_identifier = {
    name = "gen-2-8"
  }

  image = {
    id = "ubuntu-22-04-id"
  }

  network_interfaces = [
    {
      interface_id = cloudru_evolution_compute_interface.ops_interface.id
    }
  ]
}

# ==============================================================================
# 5. ВЫХОДНЫЕ ПАРАМЕТРЫ ДЛЯ ANSIBLE
# ==============================================================================
output "master_ip" {
  value = cloudru_evolution_compute_interface.master_interface.ip_address
}
output "worker_ip" {
  value = cloudru_evolution_compute_interface.worker_interface.ip_address
}
output "data_ip" {
  value = cloudru_evolution_compute_interface.data_interface.ip_address
}
output "ops_ip" {
  value = cloudru_evolution_compute_interface.ops_interface.ip_address
}