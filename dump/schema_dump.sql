-- MySQL dump 10.13  Distrib 8.0.33, for Linux (aarch64)
--
-- Host: localhost    Database: node
-- ------------------------------------------------------
-- Server version	8.0.33

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `nodes`
--

DROP TABLE IF EXISTS `nodes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `nodes` (
  `sort_order` int DEFAULT NULL,
  `computer_id` varchar(255) NOT NULL,
  `ip_address` varchar(255) DEFAULT NULL,
  `node_group` int DEFAULT NULL,
  `reachable` tinyint NOT NULL DEFAULT '1',
  PRIMARY KEY (`computer_id`),
  UNIQUE KEY `computer_id_UNIQUE` (`computer_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `nodes`
--

LOCK TABLES `nodes` WRITE;
/*!40000 ALTER TABLE `nodes` DISABLE KEYS */;
/*!40000 ALTER TABLE `nodes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `nodes_buffer`
--

DROP TABLE IF EXISTS `nodes_buffer`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `nodes_buffer` (
  `ip_address` varchar(45) NOT NULL,
  PRIMARY KEY (`ip_address`),
  UNIQUE KEY `id_UNIQUE` (`ip_address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `nodes_buffer`
--

LOCK TABLES `nodes_buffer` WRITE;
/*!40000 ALTER TABLE `nodes_buffer` DISABLE KEYS */;
/*!40000 ALTER TABLE `nodes_buffer` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `nodes_que`
--

DROP TABLE IF EXISTS `nodes_que`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `nodes_que` (
  `ip_address` varchar(45) NOT NULL,
  PRIMARY KEY (`ip_address`),
  UNIQUE KEY `idip_address_UNIQUE` (`ip_address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `nodes_que`
--

LOCK TABLES `nodes_que` WRITE;
/*!40000 ALTER TABLE `nodes_que` DISABLE KEYS */;
/*!40000 ALTER TABLE `nodes_que` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `nonce`
--

DROP TABLE IF EXISTS `nonce`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `nonce` (
  `nonce` varchar(255) NOT NULL,
  PRIMARY KEY (`nonce`),
  UNIQUE KEY `nonce_UNIQUE` (`nonce`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `nonce`
--

LOCK TABLES `nonce` WRITE;
/*!40000 ALTER TABLE `nonce` DISABLE KEYS */;
INSERT INTO `nonce` VALUES ('0eb187f2c2aa8150'),('26d2c3de977e7015'),('2dc40d7534e47e01'),('3c59bdfa46220d9b'),('57e727b31e49d3f1'),('97ccf2cf46f63f87'),('a59ae27d7fb294d9'),('c198073acb72e1bc'),('cc11c755fb9b771c'),('ce183308d40c357d'),('d91ef3f0958ddb1b'),('df4b5f895431f80e'),('ee4fd6545cbf6753');
/*!40000 ALTER TABLE `nonce` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `wallet_balances`
--

DROP TABLE IF EXISTS `wallet_balances`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `wallet_balances` (
  `wallet` varchar(255) NOT NULL,
  `balance` int DEFAULT NULL,
  PRIMARY KEY (`wallet`),
  UNIQUE KEY `wallet_UNIQUE` (`wallet`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `wallet_balances`
--

LOCK TABLES `wallet_balances` WRITE;
/*!40000 ALTER TABLE `wallet_balances` DISABLE KEYS */;
INSERT INTO `wallet_balances` VALUES ('BEt2A+KxW6ZTo06NtRRNusecPhcQaELfg8MZxqbwt+oxAAxfur+pSFiawTR6FH3Ry/QmyOOvoe7G7dTl2UsBfJ8=',50000),('BLN5Ss57+ZnqW4jKP3QuaNqT7OWHtsHzvbOpMu03tCF+nA3x7JhlO2tVnXLwHtDAg5Nf1OuNjCK41pG2pQAJx0k=',50000);
/*!40000 ALTER TABLE `wallet_balances` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-01-07  2:47:24
