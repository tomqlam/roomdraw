import React, { useContext } from "react";
import { MyContext } from "../../context/MyContext";

function TimelinePage() {
    const { setCurrPage, credentials } = useContext(MyContext);

    const body = (
        <>
            <h1 className="is-size-3 has-text-weight-semibold has-text-centered mb-5" style={{ color: "var(--text-color)" }}>
                DigiDraw Timeline Reminders!
            </h1>

            <div className="content">
                <div className="mb-5">
                    <h2 className="is-size-4 has-text-weight-semibold mb-3" style={{ color: "var(--text-color)" }}>
                        Friday, April 3
                    </h2>
                    <p className="mb-3">Digital draw officially opens at noon PDT.</p>
                    <p>
                        Seniors must participate in Digital Draw by 11:59 PM PDT on this day or receive the worst number
                        of their year.
                    </p>
                </div>

                <div className="mb-5">
                    <h2 className="is-size-4 has-text-weight-semibold mb-3" style={{ color: "var(--text-color)" }}>
                        Saturday, April 4
                    </h2>
                    <p>
                        Juniors must participate in Digital Draw by 11:59 PM PDT on this day or receive the worst number
                        of their year.
                    </p>
                </div>

                <div className="mb-5">
                    <h2 className="is-size-4 has-text-weight-semibold mb-3" style={{ color: "var(--text-color)" }}>
                        Sunday, April 5
                    </h2>
                    <p>
                        Sophomores must participate in Digital Draw by 11:59 PM PDT on this day or receive the worst
                        number of their year.
                    </p>
                </div>

                <p className="mb-4">
                    Moving to PLATT for First Day of Room Draw day on Monday, April 6th at 8 PM PDT.
                </p>
            </div>

            <div className="has-text-centered">
                <button type="button" className="button is-primary" onClick={() => setCurrPage("Home")}>
                    <span className="icon">
                        <i className="fas fa-home"></i>
                    </span>
                    <span>Back to home</span>
                </button>
            </div>
        </>
    );

    if (!credentials) {
        return (
            <section className="login-section">
                <div className="login-card about-card">{body}</div>
            </section>
        );
    }

    return (
        <section className="section">
            <div className="container" style={{ maxWidth: "820px" }}>
                <div className="box" style={{ backgroundColor: "var(--card-bg)" }}>
                    {body}
                </div>
            </div>
        </section>
    );
}

export default TimelinePage;
