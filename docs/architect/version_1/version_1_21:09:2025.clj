          +-------------------+                    +------------------+
          |   Bank System     |                    |   Core Chain     |
          | (Deposit/Withdraw)|                    | (Ethereum Testnet|
          +---------+---------+                    +---------+--------+
                    |                                          ^
       REST API     | POST /deposit                            |
   (mTLS + JWT)     v                                          |
          +-------------------+                                |
          |   Banking API     |                                |
          | (Open Banking/    |                                |
          |  PSD2 Integration)|                                |
          +---------+---------+                                |
                    |                                          |
       REST API     | POST /bridge/deposit                     |
   (mTLS + JWT)     v                                          |
          +-------------------+         Kafka        +----------+----------+
          |   Bridge API      |--------------------->|   Deposit Worker   |
          | (Go HTTP Server)  |                      | (consume events,   |
          +---------+---------+                      | sign+broadcast tx) |
                    |                                +----------+----------+
                    v                                           |
            +---------------+                                   |
            |   Postgres    |<--------------+                   |
            | (Ledger, Tx)  |               |                   |
            +-------+-------+               |                   |
                    |                       |                   |
                    v                       |                   |
             +-------------+                |                   |
             | Audit Logs  |                |                   |
             +-------------+                |                   |
                                            |                   v
                                    +-------+-------+    +-------------+
                                    |   Vault/HSM   |    | Ethereum RPC|
                                    | (sign tx)     |    | (Sepolia)   |
                                    +---------------+    +-------------+
