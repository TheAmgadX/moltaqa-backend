# User Deletion Flow

```mermaid
sequenceDiagram
    autonumber

    participant Client
    participant Gateway as API Gateway
    participant Auth as Auth Service
    participant OTP as OTP Service
    participant User as User Service
    participant UserDB as User Database
    participant Kafka
    participant Search as Search Service
    participant Social as Social Graph Service
    participant Posts as Posts Service
    participant Chat as Chat Service
    participant Notification as Notification Service

    alt Delete User

        Client->>Gateway: Request account deletion
        Gateway->>Auth: DeleteAccount()

        Auth->>OTP: Send OTP
        OTP-->>Client: Deliver OTP

        Client->>Gateway: Submit OTP
        Gateway->>Auth: Verify OTP

        Auth->>OTP: Validate OTP

        alt Invalid OTP
            OTP-->>Auth: Invalid
            Auth-->>Gateway: OTP verification failed
            Gateway-->>Client: 401 Unauthorized

        else OTP Valid

            Auth->>User: DeleteUser(userId)

            User->>UserDB: Soft delete user

            User->>Kafka: Publish UserDeleted

            par Search Service
                Kafka-->>Search: UserDeleted
                Search->>Search: Remove user from index
            and Social Graph Service
                Kafka-->>Social: UserDeleted
                Social->>Social: Hide profile
            and Posts Service
                Kafka-->>Posts: UserDeleted
                Posts->>Posts: Hide author's content
            and Chat Service
                Kafka-->>Chat: UserDeleted
                Chat->>Chat: Update cached user state
            and Notification Service
                Kafka-->>Notification: UserDeleted
            end

            User-->>Gateway: Account deleted
            Gateway-->>Client: 200 OK
        end

    end
```
