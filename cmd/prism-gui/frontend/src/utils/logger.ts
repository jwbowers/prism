/**
 * Production-safe logger utility for Prism GUI
 *
 * Provides structured logging with configurable levels and outputs to browser console.
 * This utility is eslint-safe and provides consistent logging across the application.
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LoggerConfig {
  minLevel: LogLevel;
  prefix?: string;
}

const LOG_LEVELS: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
};

class Logger {
  private config: LoggerConfig;

  constructor(config: LoggerConfig = { minLevel: 'info' }) {
    this.config = config;
  }

  private shouldLog(level: LogLevel): boolean {
    return LOG_LEVELS[level] >= LOG_LEVELS[this.config.minLevel];
  }

  private formatMessage(level: LogLevel, message: string, ...args: unknown[]): unknown[] {
    const timestamp = new Date().toISOString();
    const prefix = this.config.prefix ? `[${this.config.prefix}]` : '';
    const formattedMessage = `[${timestamp}] ${prefix}[${level.toUpperCase()}] ${message}`;
    return [formattedMessage, ...args];
  }

  debug(message: string, ...args: unknown[]): void {
    if (this.shouldLog('debug')) {
      // eslint-disable-next-line no-console
      console.debug(...this.formatMessage('debug', message, ...args));
    }
  }

  info(message: string, ...args: unknown[]): void {
    if (this.shouldLog('info')) {
      // eslint-disable-next-line no-console
      console.info(...this.formatMessage('info', message, ...args));
    }
  }

  warn(message: string, ...args: unknown[]): void {
    if (this.shouldLog('warn')) {
      // eslint-disable-next-line no-console
      console.warn(...this.formatMessage('warn', message, ...args));
    }
  }

  error(message: string, ...args: unknown[]): void {
    if (this.shouldLog('error')) {
      // eslint-disable-next-line no-console
      console.error(...this.formatMessage('error', message, ...args));
    }
  }

  log(message: string, ...args: unknown[]): void {
    // Alias for info
    this.info(message, ...args);
  }
}

// Default logger instance for the application
export const logger = new Logger({
  minLevel: process.env.NODE_ENV === 'production' ? 'warn' : 'debug',
  prefix: 'Prism',
});

// Export Logger class for custom instances if needed
export { Logger };
export type { LogLevel, LoggerConfig };
