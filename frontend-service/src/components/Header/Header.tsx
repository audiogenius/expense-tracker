import React from 'react'
import type { Profile } from '../../types'

type HeaderProps = {
  profile: Profile | null
  onLogout: () => void
}

export const Header: React.FC<HeaderProps> = ({ profile, onLogout }) => {
  return (
    <div className="header-bar">
      <div className="header-left">
        <h1 className="title" style={{ fontSize: '24px', margin: 0 }}>Expense Tracker</h1>
        {profile && (
          <div className="profile-avatar" style={{ padding: '8px 12px' }}>
            {profile.photo_url && (
              <img src={profile.photo_url} alt="avatar" style={{ width: '32px', height: '32px' }} />
            )}
            <div className="profile-info">
              <div className="profile-name">{profile.username}</div>
            </div>
          </div>
        )}
      </div>
      <button onClick={onLogout} className="logout-btn" title="Выход">
        ✕
      </button>
    </div>
  )
}

