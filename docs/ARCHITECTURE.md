# Architecture Overview

## 1. 고수준 아키텍처 (High-Level Architecture)

이 섹션에서는 프로젝트의 전체 시스템 아키텍처와 주요 구성 요소 간의 상호 작용을 설명합니다.

### 1.1. 전체 시스템 다이어그램 (Overall System Diagram)

본 프로젝트의 아키텍처는 아래 다이어그램에 요약되어 있으며, 클라이언트, 백엔드, 데이터베이스, 외부 서비스 간의 관계를 보여줍니다.

*(자세한 내용은 `docs/assets/system-architecture.puml` 파일을 참고하십시오.)*

![System Architecture Diagram](assets/system-architecture.puml)

### 1.2. 주요 구성 요소 (Key Components)

-   **Client (Flutter)**: 사용자 인터페이스를 제공하는 모바일 애플리케이션입니다. Firebase 인증을 통해 사용자를 인증하고, 백엔드 서버와 `connect-go`를 통해 통신합니다.
-   **Backend (Go, connect-go)**: Go 언어로 작성된 핵심 비즈니스 로직을 처리하는 서버입니다. `connect-go` 라이브러리를 사용하여 단일 포트에서 gRPC와 RESTful API 요청을 동시에 처리할 수 있습니다. Google Cloud Run에서 실행되도록 설계되었습니다.
-   **Databases**:
    -   **PostgreSQL (Cloud SQL)**: 영구적인 데이터를 저장하는 기본 관계형 데이터베이스입니다. 사용자 정보, 퀴즈 데이터 등 핵심 정보를 관리합니다.
    -   **Redis (Cloud Memorystore)**: 실시간 상태 저장 및 캐싱을 위한 인메모리 데이터 저장소입니다. 빠른 응답 속도가 요구되는 데이터나 임시 데이터를 처리합니다.
-   **External Services**:
    -   **Firebase (Authentication, Storage)**: 사용자 인증(ID 토큰 발급 및 검증)과 파일(예: 이미지) 저장을 위해 사용됩니다.
    -   **GCP Pub/Sub**: 비동기 메시지 처리를 위한 서비스입니다. 시스템의 다른 부분 간에 이벤트를 발행(Publish)하고 구독(Subscribe)하여 시스템의 결합도를 낮춥니다.

### 1.3. 통신 흐름 (Communication Flow)

-   **인증**: 클라이언트는 Firebase SDK를 통해 로그인하고 ID 토큰을 발급받습니다. 이후 백엔드에 보내는 모든 요청의 헤더에 이 토큰을 포함하여 자신을 인증합니다.
-   **API 통신**: 클라이언트는 `connect-go` 서버에 Unary(단일 요청/응답) 또는 Streaming(스트리밍) 방식으로 API를 요청합니다. `connect-go` 덕분에 동일한 Protobuf 정의에서 생성된 코드로 gRPC와 HTTP/1.1 (JSON) 통신을 모두 지원할 수 있습니다.
-   **비동기 처리**: 백엔드 로직 내에서 특정 이벤트가 발생하면 (예: 상태 변경), GCP Pub/Sub을 통해 메시지를 발행합니다. 관심 있는 다른 서비스나 백엔드 자체의 다른 부분이 이 메시지를 구독하여 후속 작업을 비동기적으로 처리합니다. 이는 시스템의 확장성과 탄력성을 높여줍니다.

## 2. 클린 아키텍처 (Clean Architecture)

본 프로젝트는 로버트 C. 마틴(Robert C. Martin)이 제안한 클린 아키텍처를 핵심 설계 원칙으로 채택하고 있습니다. 이를 통해 코드의 관심사를 분리하고, 테스트 용이성을 높이며, 외부 요인(프레임워크, 데이터베이스 등)의 변경에 유연하게 대처할 수 있는 구조를 지향합니다.

### 2.1. 아키텍처 원칙 및 구조

-   **의존성 규칙 (The Dependency Rule)**: 모든 소스 코드 의존성은 바깥쪽에서 안쪽으로, 즉 저수준 정책에서 고수준 정책으로 향해야 합니다. 내부 계층은 외부 계층에 대해 아무것도 알지 못합니다.
-   **계층 구조**: 코드는 크게 4개의 계층으로 나뉩니다: **Domain, Application, Data, Protocol**.
-   **기능 기반 구성 (Feature-based Organization)**: 모든 코드는 `internal/feature` 디렉토리 아래에 기능별로 그룹화됩니다. 각 기능 디렉토리(`health`, `user` 등)는 자체적인 4개의 계층 구조를 가집니다. 이를 통해 기능의 응집도를 높이고 다른 기능과의 결합도를 낮춥니다.

```
/internal
└── /feature
    ├── /health
    │   ├── /domain
    │   ├── /application
    │   ├── /data
    │   └── /protocol
    └── /user
        ├── /domain
        ├── /application
        ├── /data
        └── /protocol
```

### 2.2. 아키텍처 계층 상세 설명 (Layers in Detail)

`health` 기능을 예시로 각 계층의 역할과 상호작용을 살펴보겠습니다.

#### 2.2.1. 도메인 (Domain)

가장 안쪽의 계층으로, 시스템의 핵심 비즈니스 로직과 규칙을 포함합니다. 이 계층은 다른 어떤 계층에도 의존하지 않는 순수한 Go 코드로 작성됩니다.

-   **역할**:
    -   **엔티티 (Entities)**: 비즈니스 객체를 정의합니다. (예: `HealthStatus`)
    -   **리포지토리 인터페이스 (Repository Interfaces)**: 데이터 영속성에 대한 계약(규칙)을 정의합니다. 실제 구현은 데이터 계층에 위임됩니다. (예: `HealthRepository`)

-   **예시 (`internal/feature/health/domain/`)**:

    *엔티티 (`entity/status.go`)*
    ```go
    package entity

    // HealthStatus represents the health status of the service
    type HealthStatus struct {
        Status string `json:"status"`
    }

    func HealthyStatus() *HealthStatus {
        return NewHealthStatus("healthy")
    }
    ```

    *리포지토리 인터페이스 (`repo/repository.go`)*
    ```go
    package repo

    import (
        "context"
        "github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
    )

    // HealthRepository defines the protocol for health data access
    type HealthRepository interface {
        GetStatus(ctx context.Context) (*entity.HealthStatus, error)
        GetDatabaseStatus(ctx context.Context) (*entity.DatabaseStatus, error)
    }
    ```

#### 2.2.2. 애플리케이션 (Application)

도메인 계층을 감싸며, 시스템이 제공하는 유스케이스(사용 사례)를 구현합니다.

-   **역할**:
    -   도메인 엔티티와 리포지토리 인터페이스를 사용하여 비즈니스 로직을 오케스트레이션합니다.
    -   입력(DTO)을 받아 도메인 모델을 호출하고, 결과를 출력(DTO)으로 변환하는 흐름을 제어합니다.

-   **의존성**: 도메인 계층에만 의존합니다.

-   **예시 (`internal/feature/health/application/health_check_usecase.go`)**:
    ```go
    package application

    import (
        "context"
        "github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
        "github.com/seventeenthearth/sudal/internal/feature/health/domain/repo"
    )

    type HealthCheckUseCase interface {
        Execute(ctx context.Context) (*entity.HealthStatus, error)
    }

    type healthCheckUseCase struct {
        repo repo.HealthRepository // Domain의 인터페이스에 의존
    }

    func NewHealthCheckUseCase(repository repo.HealthRepository) HealthCheckUseCase {
        return &healthCheckUseCase{
            repo: repository,
        }
    }

    func (uc *healthCheckUseCase) Execute(ctx context.Context) (*entity.HealthStatus, error) {
        // 리포지토리를 통해 상태를 가져오는 로직을 수행
        return uc.repo.GetStatus(ctx)
    }
    ```

#### 2.2.3. 데이터 (Data)

데이터베이스, 외부 API 등 시스템 외부의 데이터 소스와의 상호작용을 담당합니다.

-   **역할**:
    -   도메인 계층에 정의된 리포지토리 인터페이스의 구체적인 구현을 제공합니다.
    -   ORM, 데이터베이스 드라이버 등 인프라 관련 기술을 사용하여 데이터를 처리합니다.

-   **의존성**: 도메인 계층에 정의된 인터페이스를 구현(implement)하며, 인프라 서비스(예: `PostgresManager`)에 의존(depend)합니다.

-   **예시 (`internal/feature/health/data/repo/repository.go`)**:
    ```go
    package repo

    import (
        "context"
        "github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
        // 인프라 서비스에 대한 의존성
        spostgres "github.com/seventeenthearth/sudal/internal/service/postgres"
    )

    // HealthRepository는 domain.repo.HealthRepository 인터페이스의 구현체
    type HealthRepository struct {
        dbManager spostgres.PostgresManager
    }

    func NewHealthRepository(dbManager spostgres.PostgresManager) *HealthRepository {
        return &HealthRepository{
            dbManager: dbManager,
        }
    }

    func (r *HealthRepository) GetStatus(ctx context.Context) (*entity.HealthStatus, error) {
        // 실제 데이터베이스 헬스 체크 로직 수행
        _, err := r.dbManager.HealthCheck(ctx)
        if err != nil {
            return entity.UnhealthyStatus(), nil
        }
        return entity.HealthyStatus(), nil
    }
    ```

#### 2.2.4. 프로토콜 (Protocol)

가장 바깥쪽 계층으로, 외부 세계와의 통신을 책임집니다.

-   **역할**:
    -   gRPC 핸들러, REST API 컨트롤러 등을 포함합니다.
    -   외부 요청(Request)을 받아 애플리케이션 계층의 유스케이스를 호출하고, 그 결과를 응답(Response)으로 변환하여 반환합니다.
    -   데이터 변환(예: 도메인 엔티티 -> Protobuf 메시지)을 담당합니다.

-   **의존성**: 애플리케이션 계층에 의존합니다.

-   **예시 (`internal/feature/health/protocol/grpc_manager.go`)**:
    ```go
    package protocol

    import (
        "context"
        "connectrpc.com/connect"
        healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
        // 애플리케이션 계층에 대한 의존성
        "github.com/seventeenthearth/sudal/internal/feature/health/application"
    )

    type HealthManager struct {
        healthService application.HealthService
    }

    func NewHealthManager(service application.HealthService) *HealthManager {
        return &HealthManager{
            healthService: service,
        }
    }

    func (h *HealthManager) Check(
        ctx context.Context,
        req *connect.Request[healthv1.CheckRequest],
    ) (*connect.Response[healthv1.CheckResponse], error) {
        // 애플리케이션 서비스 호출
        status, err := h.healthService.Check(ctx)
        if err != nil {
            // ...
        }
        // 결과를 Protobuf 메시지로 변환
        protoStatus := ToProtoServingStatus(status)
        response := &healthv1.CheckResponse{
            Status: protoStatus,
        }
        return connect.NewResponse(response), nil
    }
    ```

## 3. 공통 인프라 및 주요 패턴 (Common Infrastructure & Patterns)

이 섹션에서는 여러 기능에 걸쳐 사용되는 공통 인프라 구조와 주요 디자인 패턴에 대해 설명합니다.

### 3.1. 의존성 주입 (Dependency Injection)

본 프로젝트는 컴포넌트 간의 결합도를 낮추고 테스트 용이성을 높이기 위해 의존성 주입(DI)을 적극적으로 활용합니다. DI 컨테이너로는 Google의 [Wire](https://github.com/google/wire)를 사용합니다.

Wire는 컴파일 타임에 코드를 생성하여 의존성을 주입하므로, 런타임에 발생할 수 있는 오류를 미리 방지하고 성능 저하가 없는 장점이 있습니다. 모든 의존성 설정은 `internal/infrastructure/di/wire.go` 파일에 정의되어 있습니다.

`health` 기능의 gRPC 핸들어를 생성하는 `HealthConnectSet`을 예로 들면, 다음과 같이 각 계층의 컴포넌트들이 연결되는 것을 볼 수 있습니다.

-   **예시 (`internal/infrastructure/di/wire.go`)**:
    ```go
    // HealthConnectSet is a Wire provider set for Connect-go health service
    var HealthConnectSet = wire.NewSet(
        ProvidePostgresManager,
        // Data 계층의 구체적인 리포지토리 구현
        repo2.NewHealthRepository,
        // Domain 계층의 인터페이스와 Data 계층의 구현을 바인딩
        wire.Bind(new(repo.HealthRepository), new(*repo2.HealthRepository)),
        // Application 계층의 유스케이스
        application.NewHealthCheckUseCase,
        application.NewService,
        // Protocol 계층의 gRPC 핸들러
        healthConnect.NewHealthManager,
    )
    ```
    이처럼 Wire를 통해 `HealthManager`가 `HealthService`에, `HealthService`가 `HealthCheckUseCase`에, 그리고 `HealthCheckUseCase`가 `HealthRepository`의 구현체에 의존하는 전체 그래프가 컴파일 시점에 안전하게 구성됩니다.

### 3.2. 미들웨어 (Middleware)

인증, 로깅, 프로토콜 필터링 등 여러 API 엔드포인트에 공통으로 적용되어야 하는 기능들은 미들웨어를 통해 구현됩니다. 미들웨어는 `internal/infrastructure/middleware` 디렉토리에 위치하며, `connect-go`의 인터셉터(Interceptor)나 표준 HTTP 미들웨어 체인에 통합되어 사용됩니다.

이를 통해 API 핸들러 코드는 순수하게 비즈니스 요청 처리에만 집중할 수 있습니다.

### 3.3. 주요 디자인 패턴 (Key Design Patterns)

-   **서킷 브레이커 (Circuit Breaker)**: 시스템 아키텍처 다이어그램에 명시된 바와 같이, 외부 서비스(예: 데이터베이스) 호출 시 발생할 수 있는 장애가 전체 시스템으로 전파되는 것을 막기 위해 서킷 브레이커 패턴을 사용합니다. 특정 서비스에 대한 실패율이 임계치를 초과하면 일시적으로 해당 서비스로의 요청을 차단하여 시스템의 안정성을 유지합니다.

-   **이벤트 소싱 (Event Sourcing) 및 Pub/Sub**: 시스템의 상태 변경을 단순한 최종 상태 저장이 아닌, 발생한 이벤트의 연속으로 기록하는 이벤트 소싱 패턴의 개념이 도입되었습니다. GCP `Pub/Sub`을 활용하여 도메인 이벤트를 발행(Publish)하면, 해당 이벤트를 구독(Subscribe)하는 다른 부분에서 비동기적으로 상태를 업데이트하거나 추가 작업을 수행합니다. 이는 시스템 간의 결합도를 낮추고 확장성을 높이는 데 기여합니다.
