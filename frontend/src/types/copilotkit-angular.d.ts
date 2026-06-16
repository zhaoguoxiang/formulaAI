declare module '@copilotkitnext/angular' {
  import { Provider, Signal } from '@angular/core';

  export interface CopilotKitConfig {
    runtimeUrl?: string;
    headers?: Record<string, string>;
    licenseKey?: string;
    properties?: Record<string, unknown>;
  }

  export function provideCopilotKit(config: CopilotKitConfig): Provider;

  /** Core CopilotKit service providing runtime connection & agent management. */
  export class CopilotKit {
    readonly runtimeUrl: Signal<string>;
    readonly runtimeConnectionStatus: Signal<string>;
    readonly agents: Signal<Record<string, unknown>>;
  }
}
