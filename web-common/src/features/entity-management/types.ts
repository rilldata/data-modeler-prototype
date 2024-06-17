export enum EntityType {
  Connector = "Connector",
  Table = "Table",
  Model = "Model",
  Application = "Application",
  MetricsDefinition = "MetricsDefinition",
  MetricsExplorer = "MetricsExplorer",
  Chart = "Chart",
  Dashboard = "Dashboard",
  Unknown = "Unknown",
}

export enum StateType {
  Persistent = "Persistent",
  Derived = "Derived",
}

export interface EntityRecord {
  id: string;
  type: EntityType;
  lastUpdated: number;
}

export enum EntityStatus {
  Idle,
  Running,
  Error,

  Importing,
  Validating,
  Profiling,
  Exporting,
}
