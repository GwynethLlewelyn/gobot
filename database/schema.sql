CREATE TABLE Inventory (UUID TEXT, Name TEXT, Type TEXT, LastUpdate TEXT, Permissions TEXT);
CREATE UNIQUE INDEX InventoryIndex ON Inventory(UUID ASC, Name ASC);
CREATE TABLE Positions (PermURL TEXT, UUID TEXT, Name TEXT, OwnerName TEXT, Location TEXT, Position TEXT, Rotation TEXT, Velocity TEXT, LastUpdate TEXT, OwnerKey TEXT, ObjectType TEXT, ObjectClass TEXT, RateEnergy TEXT, RateMoney TEXT, RateHappiness TEXT);
CREATE UNIQUE INDEX PositionsIndex ON Positions(UUID ASC, Position ASC);
CREATE TABLE Obstacles (UUID TEXT, Name TEXT, BotKey TEXT, BotName TEXT, Type INTEGER, Position TEXT, Rotation TEXT, Velocity TEXT, LastUpdate TEXT, Origin TEXT, Phantom INTEGER, Prims INTEGER, BBHi TEXT, BBLo TEXT);
CREATE UNIQUE INDEX ObstaclesIndex on Obstacles (UUID ASC);
CREATE TABLE Agents (UUID TEXT, Name TEXT, OwnerName TEXT, OwnerKey TEXT, Location TEXT, Position TEXT, Rotation TEXT, Velocity TEXT, Energy TEXT, Money TEXT, Happiness TEXT, Class TEXT, SubType TEXT, PermURL TEXT, LastUpdate TEXT, BestPath TEXT, SecondBestPath TEXT, CurrentTarget TEXT);
CREATE UNIQUE INDEX AgentsIndex on Agents(OwnerKey ASC);