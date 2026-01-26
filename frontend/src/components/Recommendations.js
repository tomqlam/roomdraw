import React, { useState, useContext } from "react";
import { MyContext } from "../context/MyContext";
import BumpModal from "../modals/BumpModal";

const Recommendations = () => {
    const { isModalOpen, setIsModalOpen } = useContext(MyContext);

    const dorms = ["Atwood", "East", "Drinkward", "Linde", "North", "South", "Sontag", "West", "Case"];
    const roomTypes = ["Single", "Double", "Triple", "Quad"];
    const [selectedDorms, setSelectedDorms] = useState([
        "Atwood",
        "East",
        "Drinkward",
        "Linde",
        "North",
        "South",
        "Sontag",
        "West",
        "Case",
    ]);
    const [selectedRoomTypes, setSelectedRoomTypes] = useState(["Single", "Double", "Triple", "Quad"]);

    const recommendations = [
        {
            dormName: "Atwood",
            roomNumber: "101",
            numberOfPeople: 2,
            roomType: "Double",
        },
        {
            dormName: "Sontag",
            roomNumber: "202",
            numberOfPeople: 3,
            roomType: "Triple",
        },
        {
            dormName: "East",
            roomNumber: "303",
            numberOfPeople: 1,
            roomType: "Single",
        },
        // Add more dorms here
        {
            dormName: "Drinkward",
            roomNumber: "404",
            numberOfPeople: 4,
            roomType: "Quad",
        },
        {
            dormName: "Linde",
            roomNumber: "505",
            numberOfPeople: 2,
            roomType: "Double",
        },
        {
            dormName: "North",
            roomNumber: "606",
            numberOfPeople: 3,
            roomType: "Triple",
        },
        {
            dormName: "South",
            roomNumber: "707",
            numberOfPeople: 1,
            roomType: "Single",
        },
        {
            dormName: "West",
            roomNumber: "808",
            numberOfPeople: 2,
            roomType: "Double",
        },
        {
            dormName: "Case",
            roomNumber: "909",
            numberOfPeople: 3,
            roomType: "Triple",
        },
    ];

    const handleCheckboxChange = (value, selectedValues, setSelectedValues) => {
        if (selectedValues.includes(value)) {
            setSelectedValues(selectedValues.filter((selectedValue) => selectedValue !== value));
        } else {
            setSelectedValues([...selectedValues, value]);
        }
    };

    const handleOnClickRecommendationHandler = (e) => {
        // setCurrPage('Home');
        setIsModalOpen(true);
        // commented console.log (e);
    };

    return (
        <div>
            {isModalOpen && <BumpModal />}
            <div style={{ display: "flex", gap: "1rem", flexDirection: "row" }}>
                <h3>Dorms:</h3>
                {dorms.map((dormName, index) => (
                    <div key={index}>
                        <label className="checkbox">
                            <input
                                type="checkbox"
                                checked={selectedDorms.includes(dormName)}
                                onChange={() => handleCheckboxChange(dormName, selectedDorms, setSelectedDorms)}
                            />
                            <span style={{ marginLeft: "0.5rem" }}>{dormName}</span>
                        </label>
                    </div>
                ))}
            </div>
            <div style={{ display: "flex", gap: "1rem", flexDirection: "row" }}>
                <h3>Room Types:</h3>
                {roomTypes.map((roomType, index) => (
                    <div key={index}>
                        <label className="checkbox">
                            <input
                                type="checkbox"
                                checked={selectedRoomTypes.includes(roomType)}
                                onChange={() => handleCheckboxChange(roomType, selectedRoomTypes, setSelectedRoomTypes)}
                            />
                            <span style={{ marginLeft: "0.5rem" }}>{roomType}</span>
                        </label>
                    </div>
                ))}
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: "1rem" }}>
                {recommendations
                    .filter(
                        (recommendation) =>
                            (selectedDorms.length === 0 || selectedDorms.includes(recommendation.dormName)) &&
                            (selectedRoomTypes.length === 0 || selectedRoomTypes.includes(recommendation.roomType))
                    )
                    .map((recommendation, index) => (
                        <div className="card" key={index} onClick={handleOnClickRecommendationHandler}>
                            <div className="card-content">
                                <div className="content">
                                    <h3>{recommendation.dormName}</h3>
                                    <p>Room Number: {recommendation.roomNumber}</p>
                                    <p>Number of People in Room: {recommendation.numberOfPeople}</p>
                                    <p>Room Type: {recommendation.roomType}</p>
                                </div>
                            </div>
                        </div>
                    ))}
            </div>
        </div>
    );
};

export default Recommendations;
