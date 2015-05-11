XPUB/SUB
========




                XPUB(listener)              XSUB(dialer)
                  |                           |
                  |                     dial  |
                  |<--------------------------|
                  |                           |
                  |                    magic  |
                  |<--------------------------|
                  |                           |
                  |                  command  | AUTH | FIN | SUB
                  |<--------------------------|



device sit on WAN edge

inbound msg are all CMD's
outbound msg are data | response

### state machine
