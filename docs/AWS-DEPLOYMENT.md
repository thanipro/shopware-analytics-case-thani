# AWS Deployment Architecture

## Overview

This document outlines a production-ready deployment of the analytics platform on AWS, utilizing managed services for scalability, reliability, and operational simplicity.

## High-Level Architecture

```
┌────────────────────────────────────────────────────────────┐
│                        CloudFront                          │
│                    (Frontend CDN)                          │
└──────────────────────────┬─────────────────────────────────┘
                           │
┌──────────────────────────▼─────────────────────────────────┐
│                    Application Load Balancer               │
└───────────┬──────────────────────────┬─────────────────────┘
            │                          │
    ┌───────▼───────┐          ┌───────▼───────┐
    │  Ingestion    │          │  Analytics    │
    │  ECS Service  │          │  ECS Service  │
    └───────┬───────┘          └───────┬───────┘
            │                          │
            ▼                          │
    ┌──────────────┐                  │
    │ Amazon MSK   │                  │
    │   (Kafka)    │                  │
    └───────┬──────┘                  │
            │                          │
            ▼                          │
    ┌──────────────┐                  │
    │  Consumer    │                  │
    │ ECS Service  │                  │
    └───────┬──────┘                  │
            │                          │
            └──────────┬───────────────┘
                       ▼
            ┌──────────────────────┐
            │   Amazon RDS         │
            │   (PostgreSQL)       │
            │   Multi-AZ           │
            └──────────────────────┘
```

## Component Breakdown

### 1. Frontend (S3 + CloudFront)

**Services**:
- **S3**: Static file hosting
- **CloudFront**: Global CDN

**Configuration**:
```json
{
  "S3Bucket": {
    "BucketName": "analytics-frontend-prod",
    "WebsiteConfiguration": {
      "IndexDocument": "index.html",
      "ErrorDocument": "index.html"
    }
  },
  "CloudFrontDistribution": {
    "Origins": [{
      "DomainName": "analytics-frontend-prod.s3.amazonaws.com",
      "OriginAccessIdentity": "origin-access-identity/cloudfront/E1234567890"
    }],
    "DefaultCacheBehavior": {
      "ViewerProtocolPolicy": "redirect-to-https",
      "AllowedMethods": ["GET", "HEAD", "OPTIONS"],
      "CachedMethods": ["GET", "HEAD"],
      "Compress": true
    }
  }
}
```

**Benefits**:
- Global distribution
- Automatic scaling
- HTTPS by default
- $0.085/GB transfer

---

### 2. Ingestion Service (ECS Fargate)

**Service Configuration**:
```json
{
  "serviceName": "analytics-ingestion",
  "taskDefinition": {
    "containerDefinitions": [{
      "name": "ingestion",
      "image": "123456789.dkr.ecr.us-east-1.amazonaws.com/analytics-ingestion:latest",
      "memory": 512,
      "cpu": 256,
      "portMappings": [{
        "containerPort": 8080,
        "protocol": "tcp"
      }],
      "environment": [
        {"name": "KAFKA_BROKERS", "value": "b-1.msk-cluster.amazonaws.com:9092"},
        {"name": "LOG_LEVEL", "value": "info"}
      ],
      "secrets": [
        {"name": "KAFKA_API_KEY", "valueFrom": "arn:aws:secretsmanager:us-east-1:123:secret:kafka-key"}
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/analytics-ingestion",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }]
  },
  "desiredCount": 3,
  "deploymentConfiguration": {
    "maximumPercent": 200,
    "minimumHealthyPercent": 100
  }
}
```

**Auto-scaling**:
```json
{
  "ServiceName": "analytics-ingestion",
  "ScalableTarget": {
    "MinCapacity": 2,
    "MaxCapacity": 20
  },
  "ScalingPolicy": {
    "TargetTrackingScaling": {
      "TargetValue": 70.0,
      "PredefinedMetricType": "ECSServiceAverageCPUUtilization"
    }
  }
}
```

---

### 3. Message Queue (Amazon MSK)

**Cluster Configuration**:
```json
{
  "ClusterName": "analytics-kafka",
  "KafkaVersion": "3.5.1",
  "NumberOfBrokerNodes": 3,
  "BrokerNodeGroupInfo": {
    "InstanceType": "kafka.m5.large",
    "ClientSubnets": [
      "subnet-abc123",
      "subnet-def456",
      "subnet-ghi789"
    ],
    "StorageInfo": {
      "EbsStorageInfo": {
        "VolumeSize": 100
      }
    }
  },
  "EncryptionInfo": {
    "EncryptionInTransit": {
      "ClientBroker": "TLS",
      "InCluster": true
    }
  }
}
```

**Topic Configuration**:
```bash
kafka-topics.sh --create \
  --topic analytics-events \
  --partitions 10 \
  --replication-factor 3 \
  --config retention.ms=259200000 \
  --config compression.type=lz4
```

**Why MSK?**
- Fully managed Kafka
- Automatic patching
- Multi-AZ deployment
- Integrated with CloudWatch

**Cost**: ~$400/month (3x m5.large brokers)

---

### 4. Consumer Service (ECS Fargate)

**Service Configuration**:
```json
{
  "serviceName": "analytics-consumer",
  "taskDefinition": {
    "containerDefinitions": [{
      "name": "consumer",
      "image": "123456789.dkr.ecr.us-east-1.amazonaws.com/analytics-consumer:latest",
      "memory": 1024,
      "cpu": 512,
      "environment": [
        {"name": "KAFKA_BROKERS", "value": "b-1.msk-cluster.amazonaws.com:9092"},
        {"name": "KAFKA_GROUP_ID", "value": "analytics-consumers"},
        {"name": "DB_HOST", "value": "analytics.c1234.us-east-1.rds.amazonaws.com"},
        {"name": "DB_NAME", "value": "analytics"}
      ],
      "secrets": [
        {"name": "DB_PASSWORD", "valueFrom": "arn:aws:secretsmanager:us-east-1:123:secret:db-pass"}
      ]
    }]
  },
  "desiredCount": 5
}
```

**Auto-scaling based on Kafka lag**:
```json
{
  "ScalingPolicy": {
    "TargetTrackingScaling": {
      "CustomizedMetricSpecification": {
        "MetricName": "ConsumerLag",
        "Namespace": "Analytics",
        "Statistic": "Average"
      },
      "TargetValue": 10000
    }
  }
}
```

---

### 5. Database (Amazon RDS PostgreSQL)

**Instance Configuration**:
```json
{
  "DBInstanceIdentifier": "analytics-primary",
  "DBInstanceClass": "db.r5.xlarge",
  "Engine": "postgres",
  "EngineVersion": "15.4",
  "AllocatedStorage": 500,
  "StorageType": "gp3",
  "MultiAZ": true,
  "MasterUsername": "analytics",
  "BackupRetentionPeriod": 7,
  "PreferredBackupWindow": "03:00-04:00",
  "PreferredMaintenanceWindow": "Mon:04:00-Mon:05:00"
}
```

**Read Replica**:
```json
{
  "DBInstanceIdentifier": "analytics-replica",
  "SourceDBInstanceIdentifier": "analytics-primary",
  "DBInstanceClass": "db.r5.large",
  "PubliclyAccessible": false
}
```

**Connection Pooling (RDS Proxy)**:
```json
{
  "DBProxyName": "analytics-proxy",
  "EngineFamily": "POSTGRESQL",
  "Auth": [{
    "AuthScheme": "SECRETS",
    "SecretArn": "arn:aws:secretsmanager:us-east-1:123:secret:db"
  }],
  "RequireTLS": true,
  "IdleClientTimeout": 1800,
  "MaxConnectionsPercent": 100
}
```

**Why RDS?**
- Automated backups
- Multi-AZ failover (99.95% SLA)
- Read replicas for scaling
- Automatic patching

**Cost**: ~$400/month (r5.xlarge Multi-AZ)

---

### 6. Analytics Service (ECS Fargate)

**Service Configuration**:
```json
{
  "serviceName": "analytics-api",
  "taskDefinition": {
    "containerDefinitions": [{
      "name": "analytics",
      "image": "123456789.dkr.ecr.us-east-1.amazonaws.com/analytics-api:latest",
      "memory": 512,
      "cpu": 256,
      "portMappings": [{
        "containerPort": 8000
      }],
      "environment": [
        {"name": "DB_HOST", "value": "analytics-replica.c1234.us-east-1.rds.amazonaws.com"},
        {"name": "REDIS_HOST", "value": "analytics-cache.abc123.use1.cache.amazonaws.com"}
      ]
    }]
  },
  "desiredCount": 3
}
```

---

### 7. Caching Layer (ElastiCache Redis)

**Cluster Configuration**:
```json
{
  "CacheClusterId": "analytics-cache",
  "CacheNodeType": "cache.r5.large",
  "Engine": "redis",
  "EngineVersion": "7.0",
  "NumCacheNodes": 2,
  "AutoMinorVersionUpgrade": true,
  "SnapshotRetentionLimit": 5
}
```

**Cost**: ~$200/month (r5.large)

---

## Networking Architecture

### VPC Design

```
VPC (10.0.0.0/16)
│
├── Public Subnets (10.0.1.0/24, 10.0.2.0/24, 10.0.3.0/24)
│   └── Application Load Balancer
│   └── NAT Gateways
│
├── Private Subnets - App (10.0.11.0/24, 10.0.12.0/24, 10.0.13.0/24)
│   └── ECS Fargate Tasks (Ingestion, Consumer, Analytics)
│   └── ElastiCache
│
└── Private Subnets - Data (10.0.21.0/24, 10.0.22.0/24, 10.0.23.0/24)
    └── RDS PostgreSQL
    └── MSK Kafka
```

**Security Groups**:

```json
{
  "ALB-SG": {
    "Inbound": [
      {"Port": 443, "Source": "0.0.0.0/0"}
    ],
    "Outbound": [
      {"Port": 8080, "Destination": "Ingestion-SG"},
      {"Port": 8000, "Destination": "Analytics-SG"}
    ]
  },
  "Ingestion-SG": {
    "Inbound": [
      {"Port": 8080, "Source": "ALB-SG"}
    ],
    "Outbound": [
      {"Port": 9092, "Destination": "MSK-SG"}
    ]
  },
  "Consumer-SG": {
    "Inbound": [],
    "Outbound": [
      {"Port": 9092, "Destination": "MSK-SG"},
      {"Port": 5432, "Destination": "RDS-SG"}
    ]
  },
  "Analytics-SG": {
    "Inbound": [
      {"Port": 8000, "Source": "ALB-SG"}
    ],
    "Outbound": [
      {"Port": 5432, "Destination": "RDS-SG"},
      {"Port": 6379, "Destination": "Redis-SG"}
    ]
  },
  "RDS-SG": {
    "Inbound": [
      {"Port": 5432, "Source": "Consumer-SG"},
      {"Port": 5432, "Source": "Analytics-SG"}
    ]
  }
}
```

---

## CI/CD Pipeline

### GitHub Actions Workflow

```yaml
name: Deploy to AWS

on:
  push:
    branches: [main]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Login to ECR
        run: |
          aws ecr get-login-password --region us-east-1 | \
          docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com

      - name: Build Ingestion Service
        run: |
          cd go-ingestion
          docker build -t analytics-ingestion:${{ github.sha }} .
          docker tag analytics-ingestion:${{ github.sha }} 123456789.dkr.ecr.us-east-1.amazonaws.com/analytics-ingestion:latest
          docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/analytics-ingestion:latest

      - name: Deploy to ECS
        run: |
          aws ecs update-service \
            --cluster analytics-cluster \
            --service analytics-ingestion \
            --force-new-deployment
```

---

## Monitoring & Alerting

### CloudWatch Dashboards

**Ingestion Metrics**:
- Request count
- Error rate (4xx, 5xx)
- Latency (p50, p95, p99)
- ECS CPU/Memory

**Kafka Metrics**:
- Messages per second
- Consumer lag
- Partition distribution

**Database Metrics**:
- Connections
- Query latency
- Replication lag
- CPU utilization

### CloudWatch Alarms

```json
{
  "HighErrorRate": {
    "MetricName": "HTTPCode_Target_5XX_Count",
    "Threshold": 10,
    "EvaluationPeriods": 2,
    "AlarmActions": ["arn:aws:sns:us-east-1:123:alerts"]
  },
  "HighConsumerLag": {
    "MetricName": "ConsumerLag",
    "Threshold": 100000,
    "AlarmActions": ["arn:aws:sns:us-east-1:123:autoscale"]
  },
  "DatabaseHighCPU": {
    "MetricName": "CPUUtilization",
    "Threshold": 80,
    "AlarmActions": ["arn:aws:sns:us-east-1:123:alerts"]
  }
}
```

---

## Disaster Recovery

### Backup Strategy

**RDS**:
- Automated daily backups (7-day retention)
- Manual snapshots before major changes
- Cross-region snapshot copy

**Kafka**:
- MSK automatic backups
- Topic replication factor: 3

**Recovery Objectives**:
- **RTO** (Recovery Time Objective): 30 minutes
- **RPO** (Recovery Point Objective): 5 minutes

### Multi-Region Failover

```
Primary Region (us-east-1)        Disaster Recovery (us-west-2)
│                                 │
├── RDS Primary                   ├── RDS Replica (cross-region)
├── MSK Cluster                   ├── MSK MirrorMaker 2
├── ECS Services                  ├── ECS Services (standby)
└── S3 Bucket                     └── S3 Bucket (replicated)
```

**Failover Process**:
1. Route53 health check detects failure
2. DNS failover to us-west-2
3. Promote read replica to primary
4. Scale up ECS services

---

## Cost Estimation

### Monthly Costs (Medium Scale)

| Service | Configuration | Monthly Cost |
|---------|--------------|--------------|
| ECS Fargate (Ingestion) | 3x 0.25 vCPU, 0.5 GB | $30 |
| ECS Fargate (Consumer) | 5x 0.5 vCPU, 1 GB | $75 |
| ECS Fargate (Analytics) | 3x 0.25 vCPU, 0.5 GB | $30 |
| Amazon MSK | 3x kafka.m5.large | $400 |
| RDS PostgreSQL | db.r5.xlarge Multi-AZ | $400 |
| RDS Read Replica | db.r5.large | $200 |
| ElastiCache Redis | cache.r5.large | $200 |
| Application Load Balancer | 1 ALB | $25 |
| CloudFront | 1 TB transfer | $85 |
| S3 | 10 GB storage | $1 |
| CloudWatch Logs | 50 GB/month | $25 |
| Data Transfer | 500 GB | $45 |
| **Total** | | **~$1,516/month** |

### Cost Optimization

1. **Reserved Instances**: Save 30-50%
2. **Savings Plans**: Save 20-40%
3. **Spot Instances for Consumers**: Save 70%
4. **S3 Intelligent Tiering**: Save 40% on storage
5. **CloudFront Reserved Capacity**: Save 20%

**Optimized Cost**: ~$900/month

---

## Security Best Practices

### IAM Roles

```json
{
  "IngestionTaskRole": {
    "Effect": "Allow",
    "Action": [
      "kafka:DescribeCluster",
      "kafka:GetBootstrapBrokers",
      "kafka-cluster:Connect",
      "kafka-cluster:WriteData"
    ],
    "Resource": "arn:aws:kafka:us-east-1:123:cluster/analytics/*"
  },
  "ConsumerTaskRole": {
    "Effect": "Allow",
    "Action": [
      "kafka-cluster:ReadData",
      "secretsmanager:GetSecretValue",
      "rds-db:connect"
    ]
  }
}
```

### Encryption

- **In Transit**: TLS 1.2+ for all services
- **At Rest**:
  - RDS: AWS KMS encryption
  - MSK: AWS KMS encryption
  - S3: Server-side encryption (SSE-S3)

### Secrets Management

```bash
# Store database password
aws secretsmanager create-secret \
  --name analytics/db-password \
  --secret-string "$(openssl rand -base64 32)"

# Reference in ECS task
"secrets": [{
  "name": "DB_PASSWORD",
  "valueFrom": "arn:aws:secretsmanager:us-east-1:123:secret:analytics/db-password"
}]
```

---

## Infrastructure as Code (Terraform)

```hcl
# main.tf
module "vpc" {
  source = "./modules/vpc"
  cidr_block = "10.0.0.0/16"
}

module "msk" {
  source = "./modules/msk"
  vpc_id = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
}

module "rds" {
  source = "./modules/rds"
  vpc_id = module.vpc.vpc_id
  subnet_ids = module.vpc.database_subnets
  instance_class = "db.r5.xlarge"
}

module "ecs_ingestion" {
  source = "./modules/ecs-service"
  name = "analytics-ingestion"
  image = "123456789.dkr.ecr.us-east-1.amazonaws.com/analytics-ingestion:latest"
  cpu = 256
  memory = 512
  desired_count = 3
}
```

---

## Summary

**Deployed Architecture**:
- ✅ Highly available (Multi-AZ)
- ✅ Auto-scaling (ECS + RDS)
- ✅ Secure (VPC, SGs, encryption)
- ✅ Observable (CloudWatch)
- ✅ Cost-optimized (~$900/month)

**Next Steps**:
1. Set up Terraform/CDK infrastructure
2. Configure CI/CD pipeline
3. Implement monitoring dashboards
4. Load test and tune auto-scaling
5. Document runbooks for operations
