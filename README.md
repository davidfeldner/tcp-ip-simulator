# tcp-ip-simulator

## What are packages in your implementation? 
packages are a struck but when they are sent they are converted into a byte array
## What data structure do you use to transmit data and meta-data?
byte array
## Does your implementation use threads or processes? 
we inplemented the handshake with the net package, we use threads to create the servers and clients
## Why is it not realistic to use threads?
you are inshured that the data will be recived and the channels are reliable.
## In case the network changes the order in which messages are delivered, how would you handle message re-ordering?
if handshake if fails the client resends the first syn  
## In case messages can be delayed or lost, how does your implementation handle message loss?
if no ACK is recived the client tries to resend the packet. This is a little diffret from how the protocol works.
TCP can handle multible sends at one time and then order the packages, but because we wait for an ACK for each package.
If it is delayed and a new packages with the same ACK number is recived a ack is sent again
On the client when a ACK packet is re

## Why is the 3-way handshake important?
It ensure that both sides are ready to recive data, and establish the sequence numbers




