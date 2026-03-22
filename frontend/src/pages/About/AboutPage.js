import React, { useContext } from "react";
import { MyContext } from "../../context/MyContext";

function AboutPage() {
    const { setCurrPage, credentials } = useContext(MyContext);

    const body = (
        <>
            <h1 className="is-size-3 has-text-weight-semibold has-text-centered mb-5" style={{ color: "var(--text-color)" }}>
                Welcome to DigiDraw!
            </h1>

            <div className="content">
                <p className="mb-4">
                    (A big thank you to our great webmasters Elsa L. and Anika S. for organizing this DigiDraw website{" "}
                    <span aria-hidden="true">&lt;3</span>. Also Tom Lam and Serena Mao who first developed it!)
                </p>

                <p className="mb-4">
                    DigiDraw is meant to give students a sense of how their plans may work-out in practice. Room draw
                    mechanics as written in the Regs have been implemented into this website to replicate pulling a room
                    during Real Draw.
                </p>

                <p className="mb-4">
                    All students need to participate in DigiDraw by the deadline assigned to their class year.
                </p>

                <p className="mb-4">
                    The floorplans of DigiDraw will be open to edit at the same time for all students, and bumping other
                    students with lower priority numbers, lower class, or for other valid reasons will be allowed. Frosh
                    bumping will also be allowed. Those who have submitted gender preferences will have their preference
                    enforced when they pull into a shared living space (i.e jack&amp;jill, suite).
                </p>

                <p className="mb-4">
                    However, for the floor plans to be accurate and a helpful resource, it is highly encouraged for
                    students to update the spreadsheet regularly as plans change.
                </p>

                <p className="has-text-weight-semibold mb-5">
                    Students who do not participate in Digital Draw at least once will be assigned the worst number of
                    their year (participation is based on the Honor Code).
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
            <div className="container" style={{ maxWidth: "720px" }}>
                <div className="box" style={{ backgroundColor: "var(--card-bg)" }}>
                    {body}
                </div>
            </div>
        </section>
    );
}

export default AboutPage;
