import * as React from "react";
import {Outlet} from "react-router-dom";

export default function Desktop() {
	return (
		<div>
			<p>layout</p>
			<Outlet/>
		</div>
	)
}
