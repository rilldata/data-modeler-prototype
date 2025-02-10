export enum EntityType {
  Connector = "Connector",
  Source = "Source",
  Model = "Model",
  Table = "Table",
  Application = "Application",
  MetricsDefinition = "MetricsDefinition",
  MetricsExplorer = "MetricsExplorer",
  Chart = "Chart",
  Canvas = "Canvas",
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
