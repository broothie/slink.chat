import * as React from "react";
import {Outlet} from "react-router-dom";

export default function Layout() {
	return (
		<div className="font-serif bg-desktop-green w-screen h-screen flex relative">
			<div className="fixed bottom-5 left-5 window p-2 font-sans text-sm">
				<p>Heads up: users, channels, and messages are destroyed at the top of every hour.</p>
			</div>

			<Outlet/>
		</div>
	)
}
