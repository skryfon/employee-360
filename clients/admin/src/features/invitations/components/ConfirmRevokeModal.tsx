export function ConfirmRevokeModal({
  email,
  loading,
  error,
  onConfirm,
  onCancel,
}: {
  email: string
  loading: boolean
  error?: string
  onConfirm: () => void
  onCancel: () => void
}) {
  return (
    <div className="fixed inset-0 flex items-center justify-center bg-slate-900/50 p-4">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="revoke-title"
        className="w-full max-w-sm rounded-sm border border-slate-300 bg-white p-6"
      >
        <h2 id="revoke-title" className="text-base font-semibold text-slate-900">
          Revoke invitation
        </h2>
        <p className="mt-2 text-sm text-slate-600">
          Revoke the invitation for {email}? They will no longer be able to accept it.
        </p>
        {error && (
          <p role="alert" className="mt-2 text-xs text-red-700">
            {error}
          </p>
        )}
        <div className="mt-6 flex justify-end gap-2">
          <button
            type="button"
            onClick={onCancel}
            className="h-9 rounded-sm border border-slate-300 bg-white px-4 text-sm font-medium text-slate-800 hover:bg-slate-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={loading}
            aria-busy={loading}
            className="h-9 rounded-sm bg-red-600 px-4 text-sm font-semibold text-white hover:bg-red-700 active:bg-red-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-red-600 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-500"
          >
            Revoke
          </button>
        </div>
      </div>
    </div>
  )
}
