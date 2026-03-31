import React, { useContext } from "react";
import { MyContext } from "../../context/MyContext";

/** Served from `public/Final_Highlighted_2026_Room_Draw_Regs.pdf` */
const REGS_PDF_URL = "./Final_Highlighted_2026_Room_Draw_Regs.pdf";

/** Optional: set `REACT_APP_DIGIDRAW_TIMELINE_URL` in `.env` to use an external timeline URL instead of this site's Timeline page. */
const DIGIDRAW_TIMELINE_URL = process.env.REACT_APP_DIGIDRAW_TIMELINE_URL || "";

const PLOTS_BASE = "./plots_about/";

/** 2025 per-dorm / round statistics (see `public/plots_about`). */
const PLOTS_2025_STATS = [
    "2025-room-draw-stats-01.png",
    "2025-room-draw-stats-02.png",
    "2025-room-draw-stats-03.png",
    "2025-room-draw-stats-04.png",
    "2025-room-draw-stats-05.png",
    "2025-room-draw-stats-06.png",
    "2025-room-draw-stats-07.png",
    "2025-room-draw-stats-08.png",
];

const PLOT_2025_OVERALL = "2025-room-draw-overall.png";

/** Rising sophomore pie + bar (single screenshot). Reorder `PLOTS_2025_STATS` if plots are mis-assigned. */
const PLOT_2025_SOPHOMORE = "2025-rising-sophomore-pie-bar.png";

function plotSrc(filename) {
    return `${PLOTS_BASE}${filename}`;
}

function AboutPage() {
    const { setCurrPage, credentials } = useContext(MyContext);

    const timelineLink =
        DIGIDRAW_TIMELINE_URL !== "" ? (
            <a href={DIGIDRAW_TIMELINE_URL} target="_blank" rel="noopener noreferrer">
                DigiDraw timeline page
            </a>
        ) : (
            <a
                href="#timeline"
                onClick={(e) => {
                    e.preventDefault();
                    setCurrPage("Timeline");
                }}
                className="has-text-weight-semibold"
                style={{ color: "var(--primary-color)", textDecoration: "underline" }}
            >
                DigiDraw timeline page
            </a>
        );

    const body = (
        <>
            <h1 className="is-size-3 has-text-weight-semibold has-text-centered mb-5" style={{ color: "var(--text-color)" }}>
                Welcome to DigiDraw!
            </h1>

            <div className="content">
                <p className="mb-4">
                    (A big thank you to our great webmasters Elsa L. and Anika S. for organizing this DigiDraw website. Also Tom Lam and Serena Mao for their guidance + initial development{" "}
                    <span aria-hidden="true">&lt;3</span>.)
                </p>

                <p className="mb-4">
                    DigiDraw is meant to give students a sense of how their plans may work out in practice. Room draw
                    mechanics as written in the Regs have been implemented into this website to replicate pulling a room
                    during Real Draw.
                </p>

                <p className="mb-4">
                    All students need to participate in DigiDraw by the deadline assigned to their class year. Review
                    the {timelineLink} for the deadlines.
                </p>

                <p className="mb-4">
                    The floorplans of DigiDraw will be open to edit at the same time for all students, and bumping other
                    students with lower priority numbers, lower class year, or for other valid reasons will be allowed.
                    Frosh bumping will also be allowed. Those who have submitted gender preferences will have their
                    preference enforced when they pull into a shared living space (i.e jack&amp;jill, suite).
                </p>

                <p className="mb-4">
                    However, for the floor plans to be accurate and a helpful resource, it is highly encouraged for
                    students to update the website regularly as plans change.
                </p>

                <p className="has-text-weight-semibold mb-5">
                    Students who do not participate in Digital Draw at least once will be assigned the worst number of
                    their year (participation is based on the Honor Code).
                </p>

                <h2 className="is-size-4 has-text-weight-semibold mt-5 mb-4" style={{ color: "var(--text-color)" }}>
                    Recommended Guidelines to follow for DigiDraw
                </h2>

                <p className="mb-4">
                    Please pull into a room that is your top and most realistic choice to allow DigiDraw to be a close
                    representative of everyone&apos;s Room Draw plans.
                </p>
                <p className="mb-4">
                    The RALs hope that everyone&apos;s genuine participation will help everyone better prepare for Room
                    Draw.
                </p>
                <p className="mb-4">
                    Keep in mind that a room being available during DigiDraw does not guarantee that it will be available
                    during Real Draw.
                </p>
                <p className="mb-4">
                    Pulling as realistically as possible during DigiDraw minimizes the differences but does not eliminate
                    them.
                </p>

                <p className="mb-4">
                    Please take into consideration preferences indicated in shared living spaces when pulling into a space
                    on DigiDraw.
                </p>
                <p className="mb-4">
                    It makes DigiDraw much less accurate if you pull into a space that would be considered an invalid pull
                    during Real Draw.
                </p>
                <p className="mb-4">
                    There is a difference between gender preferences and other preferences; you are required to follow
                    gender preferences, but other preferences are non-binding.
                </p>

                <p className="mb-4">
                    Do your best to update DigiDraw when your plans change or when you get bumped from your room.
                </p>
                <p className="mb-4">
                    While each year has a deadline for when you must participate by, you are allowed and encouraged to
                    continue updating after your deadline!
                </p>

                <p className="mb-4">
                    Read the Regs for details of pulling into different dorms and different types of living spaces.{" "}
                    <a href={REGS_PDF_URL} target="_blank" rel="noopener noreferrer">
                        <i className="fas fa-file-pdf" aria-hidden="true" style={{ marginRight: "0.35rem" }}></i>
                        Final_Highlighted_2026_Room_Draw_Regs.pdf
                    </a>
                </p>

                <p className="mb-4">Please abide by the Honor Code and be respectful of each other to prevent intimidation.</p>

                <p className="mb-4">
                    Contact the RALs (
                    <a href="mailto:ral-l@g.hmc.edu">ral-l@g.hmc.edu</a>) if you have any questions, comments, or concerns.
                </p>

                <h2 className="is-size-4 has-text-weight-semibold mt-5 mb-4" style={{ color: "var(--text-color)" }}>
                    Past Room Draw Data
                </h2>

                <p className="mb-4">
                    We received a lot of requests to share more data about past Room Draws for students to gain more
                    insight into the trends of this process.
                </p>

                <h3 className="is-size-5 has-text-weight-semibold mt-4 mb-3" style={{ color: "var(--text-color)" }}>
                    2025 Room Draw Statistics
                </h3>

                <p className="mb-4">
                    Here are general graphs showing the types of rooms pulled in what round for each dorm as well as the
                    big overall room draw plot.
                </p>

                <div className="columns is-multiline is-variable is-2 mb-4">
                    {PLOTS_2025_STATS.map((name) => (
                        <div key={name} className="column is-half-tablet is-full-mobile">
                            <figure className="image about-plot-figure">
                                <img src={plotSrc(name)} alt="" className="about-plot-img" loading="lazy" />
                            </figure>
                        </div>
                    ))}
                </div>

                <figure className="image about-plot-figure mb-5">
                    <img
                        src={plotSrc(PLOT_2025_OVERALL)}
                        alt="2025 overall room draw plot"
                        className="about-plot-img"
                        loading="lazy"
                    />
                </figure>

                <h3 className="is-size-5 has-text-weight-semibold mt-4 mb-3" style={{ color: "var(--text-color)" }}>
                    2025 Room Draw Rising Sophomore Specific Plots
                </h3>

                <p className="mb-4">
                    Pie chart on the left shows what percentage of rising sophomores were pulled by upperclassmen vs pulled
                    during sophomore round. The bar graph on the right shows where rising sophomores pulled into/were
                    pulled into.
                </p>

                <figure className="image about-plot-figure mb-4">
                    <img
                        src={plotSrc(PLOT_2025_SOPHOMORE)}
                        alt="2025 rising sophomore room draw: share pulled by upperclassmen vs sophomore round, and pulls by location"
                        className="about-plot-img"
                        loading="lazy"
                    />
                </figure>
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
            <div className="container" style={{ maxWidth: "960px" }}>
                <div className="box" style={{ backgroundColor: "var(--card-bg)" }}>
                    {body}
                </div>
            </div>
        </section>
    );
}

export default AboutPage;
