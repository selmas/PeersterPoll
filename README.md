# PeersterPollParty

## Introduction

We want to create a  simple voting scheme for the nodes in the network where peers don't influence each other’s vote by publicly displaying their own vote before the result. A user can ask a question, then everyone that wants to, sends their “sealed” vote to each
other. After dispatching the votes, everyone can locally open the received messages and find out the result. By having this two phases for voting, we ensure that nobody can have even a partial result of the decision of the network.
~~To ensure that only the wanted peers are capable of voting, we add a private/public key system for authentication with a simplified GPG’s web of trust. The root key is created by the founder of the vote group and sign other public key of wanted identity, which in turn, can sign others.~~

## Goals and Functionalities

The main functionality is to allow users to propose a question to everyone and have a poll, in which the users cannot be influenced by the votes of everyone else. However, they should still have the certainty that no user changed their vote after seeing  to everyone being able to see the content of their vote without everyone noticing it. This vote can be seen as a question with several different possible answers which the proposer defines as well. Then the question and the different possible, predefined answers are spread across the network (basic Peerster functionality) and everyone in the Peerster network should get it. Then the users should be able to vote through the GUI selecting one of the answers. Once the user selects this answer the user's Peerster node will register for the vote by sending his temporary public key, signed by his master public key, to the peer that proposed the poll (registery Node).  Once the poll is closed, the registery will no longer accept new registrations for the poll and instead start to gossip the list of all the participants. Using this list the participating peers can now commit to their vote, sign this commitment using a linkable ring signature and gossip it in the network. Once a peer has received all the commits (number commit equals number of participants in poll) or a timeout has occured (to be robust against node failure of a ~~gossip this vote. However this vote will be encrypted with a key known only by the sending node. After every node received every vote from the active nodes, each node gossips their key making it possible for every node to locally calculate the vote's results.~~ ~~ The goal is not to vote anonymously but just making sure voters don't know anyone else's decision before making their own while still archiving the property that once a vote is casted, voters cannot change their mind.

**MOVE TO REPUTATION SYSTEM AND MODIFY**

In the basic set-up of the voting scheme we want to archive the following security properties for a scenario in which we have an adversary controlling one or multiple peers in the network (passive adversaries) that follows the protocol exactly but tries to learn as much as possible:

- **Privacy:** The adversary is not able to learn any information about the inputs and outputs of other peers except of what he would learn from the inputs and outputs of his corrupted peers anyway
- **Correctness:** The adversary cannot falsify the outcome of the vote
- **Fairness:** As soon as the adversary learns anything about the outcome of the vote, all honest peers will learn the complete outcome of the vote
- **Robustness:** The adversary cannot make the vote abort at all

In the second part, we will focus on active corruption. Our goal is to be able to identify maliciously acting peers in the network and use a reputation system to create incentives of behaving while preventing the detected misbehavior to influence the vote in a unallowed fashion. The different cases of malicious peers we will consider for now are (this list can be expanded as the project evolves):

- Peers that cast several different, contradicting answers for the same question (the identity of such peer is exposed by the traceable ring signature scheme when a signature is used twice)
- Peers that change another peer's vote before forwarding it

In order to identify peers that forward incorrect votes, every message will be signed by the peer casting the vote. Upon receiving a vote, each peer checks the message's integrity and authenticity.
A detected inconsistency (two different votes from the same peer A) can either mean peer A casted two different votes or a peer forwarded incorrectly. In the first case, both the vote and the peer A get excluded from the current poll. In the second case, the original vote from peer A will stay in the poll, however the malicious peer who forwarded incorrectly, as well as his vote, will be discarded from the poll.
A peer which gets excluded from a poll automatically loses some of his reputation, while all peers successfully finishing a poll gains reputation. Below a certain reputation threshold, peers will no longer be allowed to participate or create polls.


## Related Work

### Reputation System:
When it comes to reputation systems most approaches focus on two peers rating each other after an interaction. For example in an online marketplace, the user buying a product can rate the seller depending on how the experience is (the item arrived on time, promises were honored, etc.).
Another solution is seen in the incentives system of BitTorrent where a tit-for-tat scheme is used. Peers have an incentive for helping other peers by uploading parts of the file they want to download. A freeloading peer will get choked and not be able to download the full file.
We can take a similar approach to the first solution where a peer's reputation is immediately affected when another peer notices its bad behavior and rates it negatively. This will then be used to exclude peers from polls in which they misbehaved. The opposite also happens, i.e. the rewarding of well behaved peers. A peer with a good track record, will have a matching high reputation.

### Authentication Scheme:
~~For the authentication scheme, we will use a kind of web of trust, as used in GPG. It will be easier, as there is no quality of trust but only if at some point, the key signing was signed by the root key. An issue with it is that trust if recursively given to everybody who was signed by the root key. That’s not an issue per se as every participant should know the importance of signing another key, and as soon as someone get cornered into signing an attacker related key, the attacker can cross sign everything. To mitigate that, we can add revocation, should the problem arise, but having eternal key for a given origin is way easier to support. If there is any issue with a given network, we can anyway easily create a new one, with newly created key. Having a short lived voting group is pretty fit to the physical use of voting, such as during a meeting or a paraoïd group of friend.~~

## Background
We will use the same infrastructure of message distribution as the one used in Peerster, so the course’s gossip algorithm. 

~~For the encryption of voting, we will use a symmetric crypto, AES-256, because it is widely used and tested and easily implemented via the “crypto/aes” go module. For the public/private key setup, we will use the “crypto/ecdsa” go module, we do consider these two encryption scheme to be trusted, even if it was defined by FIPS, we won’t here try to get away from government validated encryption.~~

### Linkable Ring Signature 

Linkable ring signatures are a type of group signature that allow any member of a group to produce a signature on behalf of the group without revealing his identity. The property of linkability provides the signature verifier with a simple method of detecting whether a signature has been used more than once. The only requirement for this scheme is an existing public key infrastructure.

## Design and Architecture

### Voting protocol:

In the following we will shortly outline the voting protocol step-by-step. We will use the Peerster network as an underlying structure on which we will build the voting functionality.

**MODIFY**

### Round 1:
1. A peer starts a poll by setting up the QUESTION, VOTE_OPTIONS and TIME_TO_VOTE (which is defined by stating the starting time and duration of the poll), sign the poll and gossip it to all known peers
2. Upon receipt of a poll that I have not seen before (check by storing the vote ID = Key{originPeer, seqNr}), check the integrity of the poll by checking the signature and, if correct, forward it to all known peers except the one from whom it was received
3. Every peer can cast a vote by choosing one of the proposed options, encrypting it, include the vote ID, signing it and forwarding it to all known Peers
4. Upon receipt of a vote that I have not seen before (check by storing Key{votingPeer, vote ID}, Remark: this mapping from vote to voter is not a problem as the goal is not anonymity but simply to not influence other votes with your own), check the integrity of the vote by checking the signature and, if correct, store it locally before forwarding it to all known peers except the one from whom it was received

### Round 2:
Once the TIME_TO_VOTE has passed (Remark: we assume that clock are synchronized already, we can add a central time authority if needed but that doesn’t change much)
1. Upon receipt of a vote,
    - if received directly from voter, discard it
    - if seen before, gossip to reach consensus
    - else, check the integrity of the vote by checking the signature, if correct, store it locally, gossip to reach consensus

### Round 3:
Once consensus on casted votes is reached To the TAs: We struggled with deciding which algorithm would be most suited for our application.

1. Sign the symmetric key, that was used to encrypt the vote for this poll and the corresponding vote ID and gossip it
2. Upon receipt of a key that I have not seen before, check the integrity of the vote by checking the signature and, if correct, store it locally before forwarding it to all known peers except the one from whom it was received (Remark: here we assume that our gossiper network, as we created it in the course, has the property of eventual consistency)

### Round 4:
Once peer has all the symmetric keys or  timed out (to achieve robustness)
1. Use the keys to locally decrypt all the votes and locally compute the outcome of the poll


Our voting protocol has the following properties:

- **Eligibility:** only registered peers can vote, and only once. This is guaranteed by the linkable ring signature. Peers, which are not registered, cannot generate a valide signature and peers, which vote more than once, will be detected because different votes will have the same signature tag.
- **Fairness:** no early results can be obtained which could influce the remaining votes. This is guaranteed by the committment scheme. Only after every peer has committed to his vote (or a timeout occures), the votes will be opened so that every peer in the network can see the vote.
- **Vote-privacy:** it is not revealed to anyone, how a particular peer voted.This is guaranteed by the linkable ring signature. The signer stays anonymous with respect to the group of possible signers (list of participants).
- **Universal-verifiability:** any peer can locally compute the outcome of the poll, verify that only registered peers voted and that every participant casted at most one vote and 

The properties of receipt-freeness and coercioin-resistance, which are also commonly desired in the context of voting protocols, are not archieved in our protocol. The voter signes the open message of his vote with a linkable ring signature. By disclosing the voters private key which was used for this vote, anyone can compute his tag and find out which vote was signed by him. The voter can therefore prove that he voted in a certain way.

### Authentication scheme:
~~To ensure a strong authentication, we will use a kind of distributed CA, it will use a public/private key system. There is a root key, created by the started of the network, or when someone else want to create a different, not related voting group, with full trust. The nodes participating in this voting group can only do so by having it’s own public key crossed signed by the root node, or by someone which was itself signed by the root node, and so on. The public key are physically signed during a small key signing party as having a key distribution via a non-authenticated network, such a the gossiper protocol we use, is not safe. Key signing party can off course be distributed by using another trusted protocol for communication but outside of this project. This is similar to the web of trust of GPG but with only the “signed at some point by the root node” and not the quality of trusting.
The transmission itself will use the gossiper network. We consider that key signing is trust operation where a node is validated to be fit to the network. If we want stronger garanties, we can require that n person sign the key before accepting it. By having the public key distributed in the gossiper network, we ensure that even if the root node or any element of the link to a peer is down, there is no issue with it, as the trust is forwarded. For an example scenario~~

~~A want to create a new vote group, to do so, it generate a new key public private key pair and sign it’s own public key
To have some utility to the fact of voting, A ask B to come and connect, and A sign B’s public key
B knows C and want also to vote with him, so B sign C’s public key
A is bored with it, so just leave the network. There is no issue as B is trusted by the root key and C is trusted by B
E want to join but is a long term ennemi from A, so it will try to interfere with the network by sending its own vote
B and C receive E’s vote, but drop it as the signature is not trusted~~

~~To ensure that there is no replaying of message, we will use a monotone vote id per peer, which is the count of the number of vote since the beginning for this peer. As the signing operation is done on the whole message, a replay will be for an older vote thus dropped.~~

~~As every message is authenticated by some mean, if it doesn’t come from the root node or related, we just drop it. This way, we have a sybil attack free network. ~~

###Identifying malicious peers:
When a peer receives a message, it should be able to verify the message's integrity and authenticity relying on our authenticity scheme to check a digital signature. When a node identifies that a received message was tampered with, the node can suspect the forwarding node of being malicious. When this situation is identified the message should be dropped so that other nodes do not suspect a non-malicious node because it forwarded an invalid message.

Also we can detect malicious behavior when we receive two different votes authenticated by the same node. Relying on the authentication scheme can allow us to easily identify this node as malicious as the traceable signature scheme exposes the identity of a peer signing two different messages in the same round.

### Reputation system:
Each peer locally keeps a table associating peers to their reputation. The value of this reputation would be initialized to zero and would be decreased every time a peer is suspected. Therefore a negative reputation value means bad behavior. A threshold defines when a peer is assumed to be malicious and after this it is added to the blacklist, not being able to participate in any future votes. \
For each round peers have an opinion of the other peers. This means that for a certain poll, peers can either trust or suspect other peers. For every pool each node has an opinion table where they can assign either **+1** or **-1** to other peers (where **+1** means to _trust_ and **-1** means to _suspect_). This table is then shared with the other peers so everyone can compute the network's overall opinions. This is added to the value of the local reputation table which results in the updated reputation.

There are two different events which should trigger the suspicion of a peer:
- When a node receives a message that is not properly signed
- When a node receives a message and its signature is invalid, i.e. the original signed message was tampered with

(- When two different and contradicting votes are received for the same question and are authenticated by the same node, the authenticating ndoe is suspected) 

In both cases it is the forwarder of such a message that is suspected as it is trying to forward an invalid vote to the network. An honest peer drops these messages so as to not be suspected by others and there is also no advantage in forwarding them.

The reputation system comes into play when after updating the local reputation table and detecting that a peer's reputation is below the threshold. When this happens the peer in question is added to the blacklist. From this moment, this peer will not be able to participate in any future polls. A peer drops all traffic forwarded by a blacklisted node making its attempts to participate in any vote useless.

Reputations are at maximum 0. They cannot go above this value in order to prevent attackers from behaving well during a certain period of time to build up their reputation only to take advantage of this situation later on and behave maliciously.

(After the deliberation phase is over and the nodes receive the keys to open each vote, before displaying the results to the user, the reputation system comes into play. In this phase all peers share their reputation tables with the others in order for every node to understand which nodes did not behave correctly. When a node receives a table its values should be added to the local table in order for every node to have the same table locally. Nodes whose reputation goes below the threshold will be added to a blacklist and will not participate in subsequent polls. Their votes for the current question will also not be taken into account. The threshold should be calibrated so that when most users believe a peer is malicious, its activity is stopped as soon as possible.)

This is the algorithm:
- Every peer initializes its reputation table to 0
- During a vote, _trust_ or _suspect_ peers
- Every peer gossips its opinion on other peers
- Compute network's opinion with the received opinions
- Update local reputation table with network's opinion
- Identify nodes below the threshold and blacklist them


## Evaluation Plan

Conduct several attacks on the network to simulate the malicious peers behavior. The network should be able to resist these attacks, i.e. eventually disable the participation of these malicious nodes.

The possible attacks could be the following behaviors implemented in a Peerster node:

Always gossip two different contradicting votes for each question proposed
Change the content of a received vote and forward it
Create a fake vote pretending to be another node
~~We can try every message with a valid signature but a not trusted one~~
We can replay every message to see if the protocol survive it (it should)

The time requirement for a full voting from zero with N nodes should be

K    time to generate public/private key for a node (can be done in parallel)
S    time to sign public key for a node (each node can sign other when signed)
C    time to reach consensus
log(N)    best case time to broadcast to the network (depend on connectivity and luck)
O    time to open and validate vote
Total: K + S * log(N) + TIME_TO_VOTE + C + log(N) + O


