import * as React from 'react'
import TitleBar from './TitleBar'

export default function SplashWindow() {
  return (
    <div className="p-1 flex flex-col self-center mx-auto window draggable">
      <div className="draggable-handle">
        <TitleBar title="Slink" />
      </div>

      <div className="px-2 py-1 font-sans">
        <div className="bg-logo-tile flex flex-col items-center p-2">
          <div className="p-5">
            <img src="/static/logo.png" alt="Slink logo" className="w-48 h-auto" />
          </div>
          <p className="text-xl font-bold text-white">Slink Instant Messenger</p>
          <p className="text-xs text-white">Inspired by AIM and Windows 95</p>
        </div>
      </div>
    </div>
  )
}
