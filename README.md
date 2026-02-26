# Video Platform - Google Cloud Platform

A scalable, serverless video content distribution platform built with Go, Next.js, and Google Cloud Platform.

## Project Purpose

This learning and portfolio project demonstrates proficiency in:

- Google Cloud Platform architecture and best practices
- Serverless application design with Cloud Run, Firestore, and Cloud Storage
- Professional Cloud Architect certification concepts
- Modern backend development with Go
- Modern frontend development with Next.js and TypeScript
- Infrastructure as Code and deployment automation

## Architecture

The platform consists of a Next.js frontend communicating with a Go backend API hosted on Cloud Run. The backend interfaces with Firestore for metadata storage and Cloud Storage for video file persistence. Future enhancements will include Cloud CDN for content delivery and Transcoder API for video processing.

## Technologies

### Backend

- Language: Go 1.26
- Compute: Cloud Run
- Database: Firestore
- Storage: Cloud Storage
- Libraries:
  - cloud.google.com/go/firestore - Firestore SDK
  - cloud.google.com/go/storage - Cloud Storage SDK
  - github.com/joho/godotenv - Environment management

### Frontend - Planned

- Framework: Next.js 14 with App Router
- Language: TypeScript
- Deployment: Cloud Run or Firebase Hosting
- Styling: TailwindCSS

### Infrastructure

- Cloud Provider: Google Cloud Platform
- IaC: Terraform
- CI/CD: Cloud Build

## Learning Goals

This project demonstrates proficiency in the following areas:

### Cloud Architecture

- Serverless design patterns
- Multi-tier application architecture
- Cost optimization strategies
- Security best practices

### GCP Services

- Cloud Run for serverless container deployment
- Firestore for NoSQL database operations
- Cloud Storage for object storage
- Cloud CDN for content delivery
- Pub/Sub for asynchronous messaging
- Transcoder API for video processing

### Professional Cloud Architect Topics

- High availability and reliability design
- Scalability and performance optimization
- Security and compliance implementation
- Cost management and optimization
- Operational excellence practices

### Software Engineering

- Clean Architecture principles
- Separation of concerns
- Comprehensive error handling strategies
- Testing best practices
- Professional documentation standards

## Getting Started

### Prerequisites

- Go 1.26 or later
- Node.js 18 or later for frontend development
- Google Cloud SDK
- Docker for local container testing
- A GCP project with billing enabled

## Professional Cloud Architect Alignment

This project addresses the following Professional Cloud Architect exam domains:

### 1. Designing and Planning

- Designing solution infrastructure
- Planning for security and compliance
- Analyzing and optimizing technical processes
- Planning for disaster recovery

### 2. Managing Implementation

- Configuring network topologies
- Configuring individual storage systems
- Deploying solution infrastructure with automation

### 3. Ensuring Solution Success

- Monitoring and logging implementation
- Managing and provisioning solution infrastructure
- Optimizing and operating solutions

### 4. Configuring Access and Security

- Managing IAM with service accounts and minimal permissions
- Managing authentication and authorization with Firebase Auth
- Defining resource hierarchy

## Contributing

This is a learning project demonstrating GCP best practices. Feedback and suggestions are welcome through issues and pull requests.

## License

MIT License - Available for learning and portfolio purposes.

## Resources

- [GCP Documentation](https://cloud.google.com/docs)
- [Professional Cloud Architect Certification](https://cloud.google.com/certification/cloud-architect)
- [Cloud Architecture Center](https://cloud.google.com/architecture)
- [Go Best Practices](https://go.dev/doc/effective_go)
- [Effective Go Programming](https://go.dev/doc/effective_go)

## Contact

Available for questions or collaboration opportunities regarding this project.

---

Note: This is an educational project following Google Cloud recommended best practices. It is designed for deployment and use in learning environments.
