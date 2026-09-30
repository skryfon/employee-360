import type { InputHTMLAttributes, Ref } from 'react'

interface Props extends InputHTMLAttributes<HTMLInputElement> {
  label: string
  error?: string
  ref?: Ref<HTMLInputElement>
}

export function FormField({ label, error, id, ref, ...rest }: Props) {
  const inputId = id ?? rest.name
  const errId = `${inputId}-error`
  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={inputId} className="text-xs font-medium text-slate-900">
        {label}
      </label>
      <input
        id={inputId}
        ref={ref}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errId : undefined}
        className="h-9 w-full rounded-sm border border-slate-300 bg-white px-3 text-sm text-slate-900 placeholder:text-slate-500 focus:border-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-900 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-500 aria-[invalid=true]:border-red-600 aria-[invalid=true]:focus:ring-red-600"
        {...rest}
      />
      {error && (
        <p id={errId} role="alert" className="text-xs text-red-700">
          {error}
        </p>
      )}
    </div>
  )
}
