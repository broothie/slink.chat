import * as React from "react";
import {Outlet} from "react-router-dom";

export default function Layout() {
	return (
		<div className="font-serif bg-desktop-green w-screen h-screen flex">
			<Outlet/>
		</div>
	)
}
