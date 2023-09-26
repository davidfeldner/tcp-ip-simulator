# tcp-ip-simulator

## What are packages in your implementation? 
Packages are structs but when they are sent, they are converted into a byte array
## What data structure do you use to transmit data and meta-data?
byte array
## Does your implementation use threads or processes? 
We inplemented the handshake with the net package, we use threads to create the servers and clients
## Why is it not realistic to use threads?
You are insured that the data will be recived and the channels are reliable.
You can not expect that behaviour in real life
## In case the network changes the order in which messages are delivered, how would you handle message re-ordering?
If handshake fails the client resends the first syn  
## In case messages can be delayed or lost, how does your implementation handle message loss?
If no ACK is recived the client tries to resend the SYN packet 10 times waiting 1 second between the send. 
On the server if the sequence does not match no responce is sent. 
There is also a timeout such that if any message is missing or delayed the client tries to send again

## Why is the 3-way handshake important?
It ensure that both sides are ready to recive data, and establish the sequence numbers
