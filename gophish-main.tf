provider "aws" {}

// details of the aws instance
resource "aws_instance" "example" {
  ami = "ami-40d5672f"
  instance_type = "t2.micro"
  vpc_security_group_ids = ["${aws_security_group.instance.id}"]
  key_name = "${aws_key_pair.auth.id}"

  tags {
    Name = "phishing-machine"
  }

  user_data = <<HEREDOC
    #!/bin/bash
    yum update -y
    yum install wget -y
    yum install unzip -y
    su ec2-user
    cd /home/ec2-user/
    wget https://getgophish.com/releases/latest/linux/64 -O gophish-linux-64bit.zip
    unzip gophish-linux-64bit.zip
    cd gophish-linux-64bit
    sudo openssl req -newkey rsa:2048 -nodes -keyout gophish.key -x509 -days 365 -out gophish.crt -subj "/C=DE/ST=Example/L=Example/O=example/OU=Cyber"
    echo '{
      "admin_server" : {
        "listen_url" : "0.0.0.0:3333",
        "use_tls" : true,
        "cert_path" : "gophish.crt",
        "key_path" : "gophish.key"
      },
      "phish_server" : {
        "listen_url" : "0.0.0.0:8080",
        "use_tls" : false,
        "cert_path" : "example.crt",
        "key_path": "example.key"
      },
      "db_name" : "sqlite3",
      "db_path" : "gophish.db",
      "migrations_prefix" : "db/db_"
    }' > config.json
    sudo ./gophish


HEREDOC

}

// details of security groups
resource "aws_security_group" "instance" {
  name = "phishing-machine"
  description = "Phishing Campaign 2018 - Managed by Terraform"
  ingress {
    from_port = 3333
    to_port = 3333
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port = 8080
    to_port = 8080
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port = 443
    to_port = 443
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
      from_port = 0
      to_port = 0
      protocol = "-1"
      cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_key_pair" "auth" {
  key_name   = "${var.key_name}"
  public_key = "${file(var.public_key_path)}"
}

variable "public_key_path" {
  description = "Enter the path to the SSH Public Key to add to AWS."
  default     = "~/.ssh/id_rsa.pub"
}

variable "key_name" {
  default     = "example" // insert your keypair name here
  description = "Desired name of AWS key pair"
}


// outputs ip when running "terraform apply"
output "public_ip" {
  value = "${aws_instance.example.public_ip}"
}
