/**
 * Framer Motion mock for unit tests (JSDOM environment).
 * motion.* components render as plain HTML elements; AnimatePresence renders children directly.
 * This prevents animation state from interfering with DOM queries in @testing-library/react.
 */
import React from 'react';

const createMotionComponent = (tag: string) =>
  React.forwardRef(({ children, initial: _i, animate: _a, exit: _e, variants: _v, transition: _t, whileHover: _wh, whileTap: _wt, layout: _l, ...props }: any, ref: any) =>
    React.createElement(tag, { ...props, ref }, children)
  );

const motionComponents = new Proxy({} as Record<string, React.ComponentType<any>>, {
  get(target, prop: string) {
    if (!target[prop]) {
      target[prop] = createMotionComponent(prop);
    }
    return target[prop];
  },
});

export const motion = motionComponents;

export const AnimatePresence = ({ children }: { children: React.ReactNode }) => children as React.ReactElement;

export const useAnimation = () => ({ start: () => {}, stop: () => {}, set: () => {} });
export const useMotionValue = (initial: number) => ({ get: () => initial, set: () => {}, onChange: () => () => {} });
export const useTransform = (_v: any, _i: any, _o: any) => ({ get: () => 0 });
export const useSpring = (value: number) => ({ get: () => value });
export const useReducedMotion = () => false;
