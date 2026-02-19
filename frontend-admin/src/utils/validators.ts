export type Validator = (value: string) => string;

export const required = (label: string): Validator => (value: string) => {
  if (!value.trim()) {
    return `${label}不能为空`;
  }
  return '';
};

export const email = (): Validator => (value: string) => {
  const pass = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
  return pass ? '' : '邮箱格式不正确';
};

export const minNumber = (label: string, min: number): Validator => (value: string) => {
  const num = Number(value);
  if (Number.isNaN(num) || num < min) {
    return `${label}不能小于${min}`;
  }
  return '';
};

export function runValidators(value: string, rules: Validator[]): string {
  for (const rule of rules) {
    const error = rule(value);
    if (error) {
      return error;
    }
  }
  return '';
}
