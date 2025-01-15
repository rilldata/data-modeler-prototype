export type Param = string;

export type PathOption = {
  label: string;
  depth?: number;
  href?: string;
  preloadData?: boolean;
  section?: string;
  pill?: string;
  absolute?: boolean;
};

export type PathOptions = Map<Param, PathOption>;
