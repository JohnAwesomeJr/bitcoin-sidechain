Bitcoin Sidechain Project
=========================

Overview
--------

Welcome to the Bitcoin Sidechain Project! This initiative aims to create a decentralized sidechain designed to facilitate fast, day-to-day transactions using Bitcoin. By leveraging a novel architecture that maximizes incentives for node operators, we strive to maintain a highly decentralized network while ensuring swift transaction processing.

Goals
-----

The primary objectives of this project are:

1.  **High-Speed Transactions**: Achieve rapid transaction speeds suitable for everyday purchases by utilizing a group of ten randomly selected nodes to process transactions.

2.  **Decentralization**: Ensure that the network remains decentralized and permissionless, allowing equal participation and incentivizing a diverse range of node operators.

3.  **Fair Incentive Structure**: Create a system where all nodes share equally in transaction fees, promoting participation and discouraging centralization.

4.  **Verification and Security**: Implement a robust verification process among nodes, where no single node is trusted, and all transactions are cross-verified to mitigate the risk of malicious actors.

5.  **Transparent Bitcoin Certificate Supply**: Maintain a verifiable one-to-one relationship between Bitcoin certificates on the sidechain and actual Bitcoin on the main blockchain, enhancing trust and transparency.

How It Works
------------

### Node Selection and Leadership

Every four hours, a new group of ten nodes will be pseudorandomly selected based on the hash of the previous block. These nodes will take on the role of leaders responsible for processing transactions. The leadership rotation enhances decentralization by ensuring that control over transaction processing is continually distributed across the network.

### Consensus Mechanism

With only ten nodes active in each leadership group, consensus can be reached quickly, enabling faster transaction processing. Each leader node will verify the work of its peers, ensuring accountability and correctness. This collective verification means that any malicious or faulty behavior can be detected by the remaining nodes, enhancing the security of the network.

### Key Management

The selected leader nodes will hold the keys for the peg-in Bitcoin on the main Bitcoin blockchain. At the conclusion of their leadership term, the ownership of these keys will be securely transferred to the next group of ten nodes, ensuring continuous control over the assets on the sidechain.

### Decentralization and Incentives

All nodes, regardless of their leadership status, will receive a share of the transaction fees generated within the network. This fair distribution of rewards is designed to motivate more participants to operate nodes, fostering a decentralized ecosystem.

Comparison with Existing Solutions
----------------------------------

### Lightning Network

The Lightning Network is designed to enable fast transactions through payment channels between users. However, it relies on a network of hubs and often has centralized tendencies. Our project differentiates itself by focusing on a fully decentralized model where every node can participate equally in transaction processing, without reliance on central entities.

### Liquid Network

The Liquid Network, developed by Blockstream, is a federated sidechain allowing for faster transactions and issuance of tokens. However, it relies on trusted entities (the federated nodes) for its operation. In contrast, our sidechain aims to be entirely decentralized, with no single party having control over the network. By allowing any node to participate and by distributing transaction fees evenly, we enhance the network's resilience against centralization.

Current Status
--------------

This project is still in development and not yet in a finished state. We welcome contributions and feedback from the community to help refine and improve our approach. If you're interested in participating, please

INSTRUCTION FOR MIGRATIONS

- if you are in development and want to create a migration just log back into the database container using the steps above and run this cammand. ```mysqldump -u root -ptest node > /host-machine/schema_dump.sql```

- you can also do the same thing with no data in the tables like this ```mysqldump -u root -ptest --no-data node > /host-machine/schema_dump.sql```