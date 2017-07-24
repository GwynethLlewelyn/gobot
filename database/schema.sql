-- phpMyAdmin SQL Dump
-- version 4.6.6deb1+deb.cihar.com~xenial.2
-- https://www.phpmyadmin.net/
--
-- Host: localhost
-- Generation Time: Jul 24, 2017 at 08:23 PM
-- Server version: 10.0.29-MariaDB-0ubuntu0.16.04.1
-- PHP Version: 7.0.18-0ubuntu0.16.04.1

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

--
-- Database: `gobot`
--
CREATE DATABASE IF NOT EXISTS `gobot` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `gobot`;

-- --------------------------------------------------------

--
-- Table structure for table `Agents`
--

CREATE TABLE IF NOT EXISTS `Agents` (
  `UUID` char(36) NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
  `Name` varchar(256) DEFAULT NULL,
  `OwnerName` varchar(256) DEFAULT NULL,
  `OwnerKey` char(36) DEFAULT '00000000-0000-0000-0000-000000000000',
  `Location` varchar(256) DEFAULT NULL,
  `Position` varchar(256) DEFAULT NULL,
  `Rotation` varchar(256) DEFAULT NULL,
  `Velocity` varchar(256) DEFAULT NULL,
  `Energy` varchar(256) DEFAULT NULL,
  `Money` varchar(256) DEFAULT NULL,
  `Happiness` varchar(256) DEFAULT NULL,
  `Class` varchar(256) DEFAULT NULL,
  `SubType` varchar(256) DEFAULT NULL,
  `PermURL` varchar(256) DEFAULT NULL,
  `LastUpdate` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `BestPath` varchar(256) DEFAULT NULL,
  `SecondBestPath` varchar(256) DEFAULT NULL,
  `CurrentTarget` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`UUID`),
  UNIQUE KEY `AgentsIndex` (`OwnerKey`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `Inventory`
--

CREATE TABLE IF NOT EXISTS `Inventory` (
  `UUID` char(36) NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
  `Name` varchar(256) DEFAULT NULL,
  `Type` varchar(256) DEFAULT NULL,
  `LastUpdate` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `Permissions` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`UUID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `Obstacles`
--

CREATE TABLE IF NOT EXISTS `Obstacles` (
  `UUID` char(36) NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
  `Name` varchar(256) DEFAULT NULL,
  `BotKey` char(36) DEFAULT '00000000-0000-0000-0000-000000000000',
  `BotName` varchar(256) DEFAULT NULL,
  `Type` int(11) DEFAULT NULL,
  `Position` varchar(256) DEFAULT NULL,
  `Rotation` varchar(256) DEFAULT NULL,
  `Velocity` varchar(256) DEFAULT NULL,
  `LastUpdate` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `Origin` varchar(256) DEFAULT NULL,
  `Phantom` int(11) DEFAULT NULL,
  `Prims` int(11) DEFAULT NULL,
  `BBHi` varchar(256) DEFAULT NULL,
  `BBLo` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`UUID`),
  KEY `PositionIndex` (`Position`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `Positions`
--

CREATE TABLE IF NOT EXISTS `Positions` (
  `PermURL` varchar(256) DEFAULT NULL,
  `UUID` char(36) NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
  `Name` varchar(256) DEFAULT NULL,
  `OwnerName` varchar(256) DEFAULT NULL,
  `Location` varchar(256) DEFAULT NULL,
  `Position` varchar(256) DEFAULT NULL,
  `Rotation` varchar(256) DEFAULT NULL,
  `Velocity` varchar(256) DEFAULT NULL,
  `LastUpdate` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `OwnerKey` char(36) DEFAULT '00000000-0000-0000-0000-000000000000',
  `ObjectType` varchar(256) DEFAULT NULL,
  `ObjectClass` varchar(256) DEFAULT NULL,
  `RateEnergy` varchar(256) DEFAULT NULL,
  `RateMoney` varchar(256) DEFAULT NULL,
  `RateHappiness` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`UUID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `Users`
--

CREATE TABLE IF NOT EXISTS `Users` (
  `Email` varchar(128) NOT NULL DEFAULT '',
  `Password` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`Email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
