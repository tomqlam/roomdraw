// BumpFroshModal.js
import React, { useState } from 'react';
import { useContext } from 'react';
import { MyContext } from './MyContext';

const BumpFroshModal = () => {
    const {
        selectedRoomObject,
        selectedItem,
        rooms,
        activeTab,
        dormMapping,
        handleErrorFromTokenExpiry,
        setIsFroshModalOpen,
        setRefreshKey,
    } = useContext(MyContext);

    const [targetRoom, setTargetRoom] = useState("NONE");
    const [froshBumpLoading, setFroshBumpLoading] = useState(false);


    const handleSelectChange = (event) => {
        setTargetRoom(event.target.value);
        console.log("Selected room: " + event.target.value);
    };

    const handleBumpFrosh = () => {
        setFroshBumpLoading(true);
        console.log("Bumping frosh to room " + targetRoom + "from room " + selectedRoomObject.roomUUID);
        // make an api call to bump the frosh to the target room
        if (localStorage.getItem('jwt')) {
            fetch(`https://www.cs.hmc.edu/~tlam/digitaldraw/api/frosh/bump/${selectedRoomObject.roomUUID}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
                },
                body: JSON.stringify({
                    targetRoomUUID: targetRoom,
                })
            })
                .then(res => {
                    return res.json();
                })
                .then(data => {
                    console.log(data);
                    setIsFroshModalOpen(false);
                    setFroshBumpLoading(false);
                    setRefreshKey(prevKey => prevKey + 1);
                    if (handleErrorFromTokenExpiry(data)) {
                        return;
                    };
                })
                .catch(err => {
                    console.log(err);
                })
        }
    }

    return (
        <div className="modal is-active">
            <div className="modal-background"></div>
            <div className="modal-card">
                <header className="modal-card-head">
                    <p className="modal-card-title">Edit Room {selectedItem}</p>
                    <button className="delete" aria-label="close" onClick={() => setIsFroshModalOpen(false)}></button>
                </header>
                <section className="modal-card-body">
                    <label className="label">Bump these frosh to a new room</label>
                    <div className="select" style={{ marginRight: "10px" }}>
                        <select value={targetRoom} onChange={(event) => handleSelectChange(event)}>
                            <option value="">Select a frosh room</option>
                            {rooms && rooms
                                .filter(room => !room.HasFrosh && room.FroshRoomType === selectedRoomObject.froshRoomType && dormMapping[room.Dorm] === activeTab)
                                .map((room, index) => (
                                    <option key={index} value={room.RoomUUID}>Room {room.RoomID}</option>
                                ))
                            }
                        </select>
                    </div>


                </section>
                <footer className="modal-card-foot" style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <button className={`button is-primary ${froshBumpLoading ? "is-loading" : ""}`} onClick={handleBumpFrosh}>Bump these frosh!</button>

                </footer>
            </div>
        </div>
        // <div className="modal is-active">
        //     <div className="modal-background"></div>
        //     <div className="modal-content">

        //             {selectedRoomObject.hasFrosh && (
        //                 <div className="select" style={{ marginRight: "10px" }}>
        //                     <select value={selectedRoomObject} onChange={() => console.log("lol")}>
        //                         <option value="">Select a room to bump frosh to</option>
        //                         {rooms && rooms
        //                             .filter(room => !room.HasFrosh && room.FroshRoomType === selectedRoomObject.froshRoomType && dormMapping[room.Dorm] === activeTab)
        //                             .map((room, index) => (
        //                                 <option key={index} value={room.RoomID}>Room {room.RoomID}</option>
        //                             ))
        //                         }
        //                     </select>
        //                 </div>
        //             )}
        //     </div>
        //     <button className="modal-close is-large" aria-label="close"></button>
        // </div>

    );
};

export default BumpFroshModal;