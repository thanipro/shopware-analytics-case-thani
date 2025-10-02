# Deployment

How I'd deploy this to AWS in production.

## Two Main Options

### ECS Fargate

This is what I'd pick if I just want things to work without managing servers.

The Docker images are pushed to ECR, create task definitions for the backend and analytics services, and then I setup an Application Load Balancer, and let AWS handle the rest. Frontend assests can be cached with CDN providers lile cloudflare. Database switches from SQLite to RDS PostgreSQL/Clickhouse.

The nice thing is AWS manages the container orchestration - you just tell it how many tasks to run and it keeps them healthy.

### Kubernetes on EKS

I'd go this route if the team already knows Kubernetes or needs more control.

Set up an EKS cluster, write deployment manifests for the services, and use ArgoCD for GitOps deployments. This means every change to the Git repo automatically triggers a deployment - no manual steps.

Still use RDS for the database because running stateful stuff in Kubernetes is more hassle than it's worth for this use case.

## Blue-Green Deployment

The idea is to run the new version alongside the old one, then switch traffic over once you're confident it works.

For ECS, you configure the service to keep 100% capacity during deploys. It spins up new containers before killing old ones. If health checks fail on the new version, it automatically rolls back.

For Kubernetes, same concept with RollingUpdate strategy. New pods come up, old ones go down only when the new ones are healthy.

With ArgoCD it's even cleaner - update the image tag in Git, ArgoCD syncs it, Kubernetes does the rollout, and if anything fails it reverts automatically.

## Canary Deployment

Sometimes you want to be extra careful. Deploy the new version to just 10% of traffic first.

For ECS: spin up one task of the new version, configure the load balancer to send 10% of requests to it, then watch the metrics. If error rates look good after 20-30 minutes, gradually increase the percentage. If errors spike, scale it back to zero and investigate.

For Kubernetes with Istio or App Mesh: same idea but you control traffic splitting with routing rules instead of task counts.

## Feature Flags

For toggling features without redeploying, I'd use environment variables for simple cases. Update the task definition or Kubernetes deployment with the new env var value and it takes effect on the next container restart - no new build needed.

For more sophisticated control (gradual rollouts, user targeting, instant on/off), tools like LaunchDarkly or AWS AppConfig are worth it.

## Rollback

If something breaks after deployment, rolling back should be one command.

ECS: point the service back to the previous task definition version.
Kubernetes: undo the last deployment.
ArgoCD: rollback to a previous revision, or just revert the Git commit.

The Git revert approach is my favorite because it's auditable and the same workflow you already know.

## What Actually Matters

The deployment setup boils down to:
1. Get images into ECR
2. Define how to run them (task def or K8s manifest)
3. Set up health checks and auto-scaling
4. Configure alerts for when things go wrong

Everything else is just making these steps repeatable and safe.
