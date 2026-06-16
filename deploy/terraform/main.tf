# ==============================================================================
# 1. НАСТРОЙКА ПРОВАЙДЕРА CLOUD.RU ADVANCED
# ==============================================================================
terraform {
  required_providers {
    sbercloud = {
      source  = "sbercloud-terraform/sbercloud"
      version = ">= 1.10.0"
    }
  }
}

provider "sbercloud" {
  auth_url   = "https://iam.ru-moscow-1.hc.sbercloud.ru/v3" # Московский регион
  region     = "ru-moscow-1"
  access_key = var.cloud_access_key
  secret_key = var.cloud_secret_key
}

# ==============================================================================
# 2. ПЕРЕМЕННЫЕ (Входные параметры)
# ==============================================================================
variable "cloud_access_key" {
  type        = string
  description = "Access Key (из личного кабинета Cloud.ru)"
}

variable "cloud_secret_key" {
  type        = string
  description = "Secret Key (из личного кабинета Cloud.ru)"
  sensitive   = true # Скрывает пароль в логах терраформа
}

variable "existing_security_group_name" {
  type        = string
  default     = "vkr-production-sg" # Замени на имя своей группы безопасности
  description = "Имя уже созданной вручную группы безопасности"
}

variable "existing_public_ip" {
  type        = string
  default     = "123.60.208.163" # Замени на свой белый статический IP
  description = "Существующий арендованный белый IP для Master-ноды"
}

# ==============================================================================
# 3. ИСТОЧНИКИ ДАННЫХ (Существующие ресурсы в облаке)
# ==============================================================================

# Читаем настройки существующей группы безопасности
data "sbercloud_networking_secgroup" "existing_sg" {
  name = var.existing_security_group_name
}

# Читаем настройки существующего белого IP адреса
data "sbercloud_vpc_eip" "existing_master_eip" {
  public_ip = var.existing_public_ip
}

# ==============================================================================
# 4. СОЗДАНИЕ СЕТИ (VPC и Подсеть)
# ==============================================================================

# Создаем Виртуальное Частное Облако (VPC)
resource "sbercloud_vpc" "vkr_vpc" {
  name = "vkr-payment-vpc"
  cidr = "10.0.0.0/16"
}

# Создаем подсеть 10.0.1.0/24 внутри VPC
resource "sbercloud_vpc_subnet" "vkr_subnet" {
  name       = "vkr-subnet"
  vpc_id     = sbercloud_vpc.vkr_vpc.id
  cidr       = "10.0.1.0/24"
  gateway_ip = "10.0.1.1"
}

# ==============================================================================
# 5. СОЗДАНИЕ ВИРТУАЛЬНЫХ МАШИН (ECS)
# ==============================================================================

# VM-1: Master / Control Plane (2 vCPU, 4GB RAM)
resource "sbercloud_compute_instance" "vm_master" {
  name              = "vm-1-master"
  image_name        = "Ubuntu 22.04"
  flavor_id         = "s2.medium.2" # 2 vCPU, 4GB RAM
  security_groups   = [data.sbercloud_networking_secgroup.existing_sg.name]
  availability_zone = "ru-moscow-1a"

  network {
    uuid = sbercloud_vpc_subnet.vkr_subnet.id
  }
}

# VM-2: Worker / Apps Node (4 vCPU, 8GB RAM)
resource "sbercloud_compute_instance" "vm_worker" {
  name              = "vm-2-worker"
  image_name        = "Ubuntu 22.04"
  flavor_id         = "s2.large.2" # 4 vCPU, 8GB RAM
  security_groups   = [data.sbercloud_networking_secgroup.existing_sg.name]
  availability_zone = "ru-moscow-1a"

  network {
    uuid = sbercloud_vpc_subnet.vkr_subnet.id
  }
}

# VM-3: Data Node (4 vCPU, 8GB RAM)
resource "sbercloud_compute_instance" "vm_data" {
  name              = "vm-3-data"
  image_name        = "Ubuntu 22.04"
  flavor_id         = "s2.large.2" # 4 vCPU, 8GB RAM
  security_groups   = [data.sbercloud_networking_secgroup.existing_sg.name]
  availability_zone = "ru-moscow-1a"

  network {
    uuid = sbercloud_vpc_subnet.vkr_subnet.id
  }
}

# VM-4: Ops / Monitoring Node (4 vCPU, 8GB RAM)
resource "sbercloud_compute_instance" "vm_ops" {
  name              = "vm-4-ops"
  image_name        = "Ubuntu 22.04"
  flavor_id         = "s2.large.2" # 4 vCPU, 8GB RAM
  security_groups   = [data.sbercloud_networking_secgroup.existing_sg.name]
  availability_zone = "ru-moscow-1a"

  network {
    uuid = sbercloud_vpc_subnet.vkr_subnet.id
  }
}

# ==============================================================================
# 6. СВЯЗЫВАНИЕ БЕЛОГО IP С MASTER-НОДОЙ
# ==============================================================================
resource "sbercloud_compute_eip_associate" "associate_master" {
  public_ip   = data.sbercloud_vpc_eip.existing_master_eip.address
  instance_id = sbercloud_compute_instance.vm_master.id
}

# ==============================================================================
# 7. ПОСТОЯННЫЙ ОБЛАЧНЫЙ ДИСК ДЛЯ БАЗЫ ДАННЫХ (EVS)
# ==============================================================================

# Создаем диск EVS на 20 ГБ
resource "sbercloud_evs_volume" "postgres_disk" {
  name              = "vkr-postgres-persistent-disk"
  availability_zone = "ru-moscow-1a" # Должна совпадать с зоной VM-3
  volume_type       = "SAS"             # Надежный и стандартный тип диска
  size              = 20
}

# Прикрепляем созданный диск к VM-3 (Data Node)
resource "sbercloud_compute_volume_attach" "postgres_disk_attach" {
  instance_id = sbercloud_compute_instance.vm_data.id
  volume_id   = sbercloud_evs_volume.postgres_disk.id
}

# ==============================================================================
# 8. ВЫХОДНЫЕ ПАРАМЕТРЫ (Облегчают настройку Ansible)
# ==============================================================================

output "vm_1_master_public_ip" {
  value       = data.sbercloud_vpc_eip.existing_master_eip.public_ip
  description = "Публичный IP адрес для подключения по SSH"
}

output "vm_1_master_private_ip" {
  value       = sbercloud_compute_instance.vm_master.access_ip_v4
  description = "Внутренний IP-адрес Master-ноды"
}

output "vm_2_worker_private_ip" {
  value       = sbercloud_compute_instance.vm_worker.access_ip_v4
  description = "Внутренний IP-адрес Worker-ноды (Apps)"
}

output "vm_3_data_private_ip" {
  value       = sbercloud_compute_instance.vm_data.access_ip_v4
  description = "Внутренний IP-адрес Data-ноды (Postgres + Kafka)"
}

output "vm_4_ops_private_ip" {
  value       = sbercloud_compute_instance.vm_ops.access_ip_v4
  description = "Внутренний IP-адрес Ops-ноды (Monitoring)"
}