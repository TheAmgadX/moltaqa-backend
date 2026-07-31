# Verify OTP Flow done in the Auth Service

```mermaid
sequenceDiagram
    autonumber

    participant Client
    participant Gateway as API Gateway
    participant Auth as Auth Service
    participant AuthCache as Auth Cache
    participant User as User Service
    participant UserDB as User Database
    participant Kafka
    participant Search as Search Service
    participant Social as Social Graph Service
    participant Posts as Posts Service
    participant Chat as Chat Service
    participant Notification as Notification Service

    %% --- OTP VERIFICATION REQUEST ---
    Client->>Gateway: Submit OTP verification request (REST API)
    Gateway->>Auth: VerifyOTP(phone/email, otp)

    %% --- OTP & TRANSACTION VALIDATION ---
    Auth->>AuthCache: Get OTP_Transaction by key (email/phone)
    AuthCache-->>Auth: Return OTP_Transaction {email/phone, action}

    alt Invalid OTP or Transaction Expired
        Auth-->>Gateway: OTP verification failed
        Gateway-->>Client: 401 Unauthorized
    else Valid OTP

        %% --- STATE MACHINE: ACTION EXECUTION ---
        alt Action == "Login" / "Register"
            Auth->>User: gRPC CheckUserExists(phone/email)
            User->>UserDB: Query user record
            UserDB-->>User: User existence status
            User-->>Auth: User exists boolean

            alt User does not exist
                Auth->>User: gRPC CreateUser(phone/email)
                User->>UserDB: Insert user record
                User-->>Auth: User created successfully
            end

        else Action == "Restore"
            Auth->>User: gRPC RestoreUser(userId)
            User->>UserDB: Restore user record
            User->>Kafka: Publish UserRestored

            par Downstream Service Sync (Restore)
                Kafka-->>Search: UserRestored (Re-index user)
            and
                Kafka-->>Social: UserRestored (Restore profile visibility)
            and
                Kafka-->>Posts: UserRestored (Restore author's content)
            and
                Kafka-->>Chat: UserRestored (Refresh cached user state)
            and
                Kafka-->>Notification: UserRestored
            end

        else Action == "Delete"
            Auth->>User: gRPC DeleteUser(userId)
            User->>UserDB: Delete/Soft-delete user record
            User->>Kafka: Publish UserDeleted

            par Downstream Service Sync (Delete)
                Kafka-->>Search: UserDeleted (Remove user from index)
            and
                Kafka-->>Social: UserDeleted (Hide profile)
            and
                Kafka-->>Posts: UserDeleted (Hide author's content)
            and
                Kafka-->>Chat: UserDeleted (Update cached user state)
            and
                Kafka-->>Notification: UserDeleted
            end
        end

        %% --- TRANSACTION CLEANUP & TOKEN ISSUANCE ---
        Auth->>AuthCache: Invalidate/Delete OTP_Transaction
        Auth-->>Gateway: Return JWT Token & Operation Status
        Gateway-->>Client: 200 OK (JWT Token)
    end
```
