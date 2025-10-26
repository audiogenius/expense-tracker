/// <reference types="react" />
/// <reference types="react-dom" />

declare module 'react' {
  export function useState<T>(initialState: T | (() => T)): [T, (value: T | ((prev: T) => T)) => void];
  export function useEffect(effect: () => void | (() => void), deps?: any[]): void;
  export function useRef<T>(initialValue: T): { current: T };
  export const FC: React.FunctionComponent;
  export interface FunctionComponent<P = {}> {
    (props: P): React.ReactElement | null;
  }
  export interface ReactElement {
    type: any;
    props: any;
    key: string | number | null;
  }
  export namespace React {
    interface IntrinsicElements {
      [elemName: string]: any;
    }
    interface FunctionComponent<P = {}> {
      (props: P): ReactElement | null;
    }
  }
  export default any;
}

declare module 'react-dom' {
  export * from 'react-dom';
}
