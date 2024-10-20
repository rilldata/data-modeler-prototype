import type { ScaleLinear, ScaleTime } from "d3-scale";
import type { Readable } from "svelte/store";

export interface ExtremumResolutionTweenProps<T> {
  delay?: number;
  duration?: number | ((from: T, to: T) => number);
  easing?: (t: number) => number;
  interpolate?: (a: T, b: T) => (t: number) => T;
}

export interface ExtremumResolutionStore extends Readable<number | Date> {
  setWithKey: (arg0: string, arg1: number | Date, arg2?: boolean) => void;
  removeKey: (arg0: string) => void;
  setTweenProps: (arg0: ExtremumResolutionTweenProps<number>) => void;
}

export type GraphicScale =
  | ScaleLinear<number, number>
  | ScaleTime<Date, number>;
export interface ScaleStore extends Readable<GraphicScale> {
  type: string;
}

export interface SimpleDataGraphicConfigurationArguments {
  id: string;
  width: number;
  height: number;
  left: number;
  right: number;
  top: number;
  bottom: number;
  fontSize: number;
  textGap: number;
  xType: ScaleType;
  yType: ScaleType;
  xMin: number | Date;
  xMax: number | Date;
  yMin: number | Date;
  yMax: number | Date;
  bodyBuffer: number;
  marginBuffer: number;
  devicePixelRatio: number;
}

export interface SimpleGraphicConfigurationDerivations {
  bodyLeft: number;
  bodyRight: number;
  bodyTop: number;
  bodyBottom: number;
  plotLeft: number;
  plotRight: number;
  plotTop: number;
  plotBottom: number;
  graphicWidth: number;
  graphicHeight: number;
}

export interface SimpleDataGraphicConfiguration
  extends SimpleDataGraphicConfigurationArguments,
    SimpleGraphicConfigurationDerivations {}

export interface CascadingContextStore<Arguments, StateStructure>
  extends Readable<StateStructure> {
  hasParentCascade: boolean;
  reconcileProps: (props: Arguments) => void;
}

export type SimpleConfigurationStore = CascadingContextStore<
  SimpleDataGraphicConfigurationArguments,
  SimpleDataGraphicConfiguration
>;
