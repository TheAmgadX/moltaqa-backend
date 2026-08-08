# Email Service Sequence Diagram

``` mermaid
sequenceDiagram
    participant Auth as Auth Service
    participant KEmail as Kafka (SendEmail)
    participant KSMS as Kafka (SendSMS)
    participant ES as Email Service
    participant SMTP as SMTP Provider

    Auth->>KEmail: Publish SendEmail event
    Auth->>KSMS: Publish SendSMS event

    KEmail-->>ES: Consume SendEmail
    ES->>SMTP: Send Email
    SMTP-->>ES: Delivery Result

    KSMS-->>ES: Consume SendSMS
    Note over ES: Validate and log SMS event\nNo SMS provider configured
```
