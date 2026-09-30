import { Link, useNavigate } from 'react-router-dom'
import { InviteUserForm } from '../components/InviteUserForm'

export default function CreateInvitationPage() {
  const navigate = useNavigate()
  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-4">
      <Link
        to="/invitations"
        className="inline-flex h-9 w-fit items-center gap-2 rounded-sm border border-slate-300 bg-white px-4 text-sm font-medium text-slate-800 hover:bg-slate-100 active:bg-slate-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2"
      >
        <svg aria-hidden="true" viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M19 12H5" />
          <path d="m12 19-7-7 7-7" />
        </svg>
        Back to invitations
      </Link>
      <h1 className="text-xl font-semibold text-slate-900">Create invitation</h1>
      <InviteUserForm onSuccess={() => navigate('/invitations')} onCancel={() => navigate('/invitations')} />
    </div>
  )
}
