import React, { useState, useContext, useEffect } from 'react';
import { MyContext } from './MyContext';
import { jwtDecode } from "jwt-decode";
import Select from "react-select";


function BumpModal() {
  const {
    selectedItem,
    credentials,
    print,
    setIsModalOpen,
    selectedOccupants,
    setSelectedOccupants,
    pullMethod,
    setPullMethod,
    showModalError,
    setShowModalError,
    getNameById,
    userMap,
    setRefreshKey,
    refreshKey,
    selectedRoomObject,
    selectedSuiteObject,
    pullError,
    setPullError,
    activeTab,
    rooms,
    setCredentials,
    handleErrorFromTokenExpiry,
    dormMapping,
    getRoomUUIDFromUserID

  } = useContext(MyContext);

  // List of arrays with two elements, where the first element is the occupant ID and the second element is the room UUID
  const [peopleWhoCanPullSingle, setPeopleWhoCanPullSingle] = useState([["Example ID", "Example Room UUID"]]);
  const [roomsWhoCanAlternatePull, setRoomsWhoCanAlternatePull] = useState([["Example Room Number", "Example Room UUID"]]);
  const [peopleAlreadyInRoom, setPeopleAlreadyInRoom] = useState([]); // list of numeric IDs of people already in the Room
  const [loadingSubmit, setLoadingSubmit] = useState(false);
  const [loadingClearPerson, setLoadingClearPerson] = useState([]);
  const [loadingClearRoom, setLoadingClearRoom] = useState(false);

  useEffect(() => {
    // If the selected suite or room changes, change the people who can pull 
    // commented console.log (selectedRoomObject.froshRoomType);
    // commented console.log (rooms);
    if (selectedSuiteObject) {
      const otherRooms = selectedSuiteObject.rooms;
      const otherOccupants = [];
      const otherRoomsWhoCanAlternatePull = [];
      for (let room of otherRooms) {
        if (room.roomNumber !== selectedItem && room.maxOccupancy === 1 && room.occupant1 !== 0 && room.pullPriority.pullType === 1) {
          otherOccupants.push([room.occupant1, room.roomUUID]);
        }
        if (selectedRoomObject.maxOccupancy === 2 && room.roomNumber !== selectedItem && room.maxOccupancy === 2) {
          otherRoomsWhoCanAlternatePull.push([room.roomNumber, room.roomUUID]);
        }

      }
      //// commented console.log (otherRoomsWhoCanAlternatePull);
      setRoomsWhoCanAlternatePull(otherRoomsWhoCanAlternatePull);
      setPeopleWhoCanPullSingle(otherOccupants);
    }
  }, [selectedSuiteObject, selectedItem]);

  function postToFrosh(roomObject) {
    fetch(`/frosh/${roomObject.roomUUID}`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
      },
    })
      .then(response => response.json())
      .then(data => {
        // commented console.log (data);
        closeModal();
        setRefreshKey(prev => prev + 1);
        if (handleErrorFromTokenExpiry(data)) {
          return;
        };
      })
      .catch((error) => {
        console.error('Error:', error);
      });
  }




  const handlePullMethodChange = (e) => {
    print(pullMethod);
    setPullMethod(e.target.value);
  };
  const closeModal = () => {
    setLoadingSubmit(false);
    setLoadingClearPerson(loadingClearPerson.map((person) => false));
    setLoadingClearRoom(false);
    setShowModalError(false);
    setPullError("");
    setIsModalOpen(false);
  };
  const handleDropdownChange = (index, value) => {
    print(value);
    const updatedselectedOccupants = [...selectedOccupants];
    updatedselectedOccupants[index - 1] = value;
    setSelectedOccupants(updatedselectedOccupants);
    print(selectedOccupants);
    setPeopleAlreadyInRoom([]);
    setShowModalError(false);
    setPullError("");

  };

  const handleSubmit = async (e) => {  // Declare handleSubmit as async
    // Handle form submission logic here
    setLoadingSubmit(true);
    e.preventDefault();
    if (pullMethod.startsWith("Alt Pull")) {
      let roomUUID = pullMethod.slice("Alt Pull ".length).trim();
      if (await canIAlternatePull(roomUUID)) {  // Wait for canIBePulled to complete
        print("This room was successfully alternative pulled");
        closeModal();
      } else {
        setLoadingSubmit(false);
        setShowModalError(true);
      }
    }
    else if (/^\d+$/.test(pullMethod)) {
      // commented console.log ("Pull method is a number");
      // pullMethod only includes number, implying that you were pulled by someone else
      if (await canIBePulled()) {  // Wait for canIBePulled to complete
        print("This room was successfully pulled by someone else in the suite");
        closeModal();
      } else {
        setLoadingSubmit(false);
        setShowModalError(true);
      }
    } else if (pullMethod === "Lock Pull") {
      // lock pulled 
      if (await canILockPull()) {  // Wait for canIBePulled to complete
        print("This room was successfully lock pulled");
        closeModal();
      } else {
        setLoadingSubmit(false);
        setShowModalError(true);
      }
    } else if (pullMethod === "Alternate Pull") {
      // Pulled with 2nd best number of this suite
      // if (await canIAlternatePull()) {  // Wait for canIBePulled to complete
      //   print("This room was successfully pulled with 2nd best number of this suite");
      //   closeModal();
      // } else {
      //   setLoadingSubmit(false);
      //   setShowModalError(true);
      // }
    } else {
      // pullMethod is either Lock Pull or Pulled themselves 
      if (await canIBump()) {  // Wait for canIBePulled to complete

        print("This room pulled themselves");
        closeModal();
        setRefreshKey(prev => prev + 1);
      } else {
        setLoadingSubmit(false);
        setShowModalError(true);
      }

    }


  };

  const performRoomAction = (pullType, pullLeaderRoom = null) => {
    // commented console.log ("performing room actoin");
    setLoadingSubmit(true);
    return new Promise((resolve) => {
      fetch(`/rooms/${selectedRoomObject.roomUUID}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
        },
        body: JSON.stringify({
          proposedOccupants: selectedOccupants
            .filter(occupant => occupant !== '')
            .map(occupant => Number(occupant.value)),
          pullType,
          pullLeaderRoom,
        }),
      })
        .then(response => response.json())
        .then(data => {
          if (data.error) {
            if (handleErrorFromTokenExpiry(data)) {
              return;
            };
            if (data.error === "One or more of the proposed occupants is already in a room") {
              // commented console.log ("Someone's already there rrror:");
              // commented console.log (data.occupants);

              // But wait: if these are the same people in the current room, handle clearing the room in the backend
              // Check if all occupants are in the current room
              if (data.occupants.length !== 0 && data.occupants.every(occupant => selectedRoomObject.roomUUID === getRoomUUIDFromUserID(occupant))) {
                // Clear the room and retry
                print("That's the case!");
                setLoadingSubmit(true);
                handleClearRoom(selectedRoomObject.roomUUID, false, -1)
                  .then(() => performRoomAction(pullType, pullLeaderRoom))
                  .then((data) => {
                    // commented console.log (data);
                    if (data === true) {
                      setRefreshKey(prev => prev + 1);
                      closeModal();
                      // commented console.log ("THERE HAS NOT BEEN A ERROR");

                    } else {
                      setLoadingSubmit(false);
                      setShowModalError(true);
                      // commented console.log ("THERE HAS BEEN A ERROR");
                      resolve(false);
                    }
                    resolve(true);
                  });
                return;
              } else {
                setPeopleAlreadyInRoom((data.occupants));
                setLoadingClearPerson(data.occupants.map((person) => false));

                const names = data.occupants.map(getNameById).join(', ');
                setPullError("Please remove " + names + " from their existing room");

              }




            } else {
              setPullError(data.error);
            }
            setIsModalOpen(true);
            setShowModalError(true);
            resolve(false);
          } else {
            // commented console.log ("Refreshing and setting");
            setRefreshKey(prev => prev + 1);
            resolve(true);
          }
        })
        .catch((error) => {
          console.error(error.error);
          setRefreshKey(prev => prev + 1);
          resolve(true);
        });
    });
  }

  const canIBump = () => performRoomAction(1);
  const canIBePulled = () => performRoomAction(2, peopleWhoCanPullSingle.find(person => person[0] === Number(pullMethod))[1]);
  const canILockPull = () => performRoomAction(3);
  const canIAlternatePull = (roomUUID) => {
    // const otherRoomInSuite = selectedSuiteObject.rooms.find(room => room.roomUUID !== selectedRoomObject.roomUUID);
    return performRoomAction(4, roomUUID);
    // if (otherRoomInSuite) {
    //   // commented console.log ("Successfully found other room");
    //   return performRoomAction(4, otherRoomInSuite.roomUUID);
    // } else {
    //   // commented console.log ("No other room in suite, can't alternate pull");
    //   return false;
    // }
  };

  const handleClearRoom = (roomUUID, closeModalBool, personIndex) => {
    return new Promise((resolve) => {
      fetch(`/rooms/${roomUUID}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('jwt')}`
        },
        body: JSON.stringify({
          proposedOccupants: [],
          pullType: 1,
        }),
      })
        .then(response => {
          // commented console.log ("Fetch response status:", response.status);  // Add this line
          return response;
        })
        .then(data => {
          // commented console.log ("Received response from clearing room");
          if (data.error) {
            if (handleErrorFromTokenExpiry(data)) {
              return;
            };
            setPullError(data.error);
            setIsModalOpen(true);
            setShowModalError(true);
            resolve(false);


          } else {
            // no error 
            // commented console.log ("Refreshing and settingloadClearPeron");
            setRefreshKey(prev => prev + 1);
            resolve(true);
            if (personIndex !== -1) {
              setLoadingClearPerson(loadingClearPerson.filter((_, itemIndex) => itemIndex !== personIndex));
              setPeopleAlreadyInRoom(peopleAlreadyInRoom.filter((_, itemIndex) => itemIndex !== personIndex));
              setShowModalError("");
            }
            if (closeModalBool) {
              closeModal();
            }


          }

        })
        .catch((error) => {
          setRefreshKey(prev => prev + 1);
          resolve(true);
          if (closeModalBool) {
            closeModal();
          } else {
            // On tap action for clearing other room
            // Also clear the error and the button 
            setShowModalError(false);
            setPullError("");
            setPeopleAlreadyInRoom([]);
          }

        });
    });
  }


  return (

    <div className="modal is-active">
      <div className="modal-background"></div>
      <div className="modal-card">
        <header className="modal-card-head">
          <p className="modal-card-title">
            {selectedRoomObject.pullPriority.isPreplaced ? "Can't edit preplaced room" : `Edit Room ${selectedItem}`}
          </p>
          <button className="delete" aria-label="close" onClick={closeModal}></button>
        </header>
        <section className="modal-card-body">


          {((jwtDecode(credentials).email === "tlam@g.hmc.edu") || (jwtDecode(credentials).email === "smao@g.hmc.edu")) && <button onClick={() => postToFrosh(selectedRoomObject)}>Add Frosh</button>}


          {!selectedRoomObject.pullPriority.isPreplaced && <div>
            <div>
              <label className="label">{`Reassign Occupant${selectedRoomObject.maxOccupancy > 1 ? "s" : ""}`}</label>

              {[1, 2, 3, 4].slice(0, selectedRoomObject.maxOccupancy).map((index) => (
                <div className="field" key={index}>
                  <div className="control">
                    <div style={{ marginBottom: "10px", width: 200 }}>
                      <Select
                        placeholder={`Select Occupant ${index}`}
                        value={
                          selectedOccupants[index - 1]
                        }
                        menuPortalTarget={document.body}
                        styles={{
                          menuPortal: base => ({ ...base, zIndex: 9999 }),
                          option: (provided, state) => ({
                            ...provided,
                            // color: 'red',
                            // backgroundColor: 'blue'
                          }),
                        }}
                        onChange={(selectedOption) => handleDropdownChange(index, selectedOption)}
                        options={userMap && Object.keys(userMap)
                          .sort((a, b) => {
                            const nameA = `${userMap[a].FirstName} ${userMap[a].LastName}`;
                            const nameB = `${userMap[b].FirstName} ${userMap[b].LastName}`;
                            return nameA.localeCompare(nameB);
                          })
                          .filter((key) => Number(userMap[key].Year) !== 0) // Replace 'YourCondition' with the condition you want to check
                          .map((key) => ({
                            value: key,
                            label: `${userMap[key].FirstName} ${userMap[key].LastName}`
                          }))}
                        />
                    </div>
                  </div>
                </div>
              ))}
            </div>
            <div>
              <label className="label" >How did they pull this room?</label>
              <div className="select">
                <select value={pullMethod} onChange={handlePullMethodChange}>
                  <option value="Pulled themselves">Pulled themselves</option>
                  {selectedRoomObject.maxOccupancy === 1 && peopleWhoCanPullSingle.map((item, index) => (
                    <option key={index} value={item[0]}>
                      Pulled by {getNameById(item[0])}
                    </option>
                  ))}
                  {selectedSuiteObject.alternative_pull && roomsWhoCanAlternatePull.map((room, index) => (
                    <option key={index} value={`Alt Pull ${room[1]}`}>Pull w/ 2nd best of {selectedRoomObject.roomNumber} and {room[0]}</option>
                  ))}
                  {selectedSuiteObject.can_lock_pull && <option value="Lock Pull">Lock Pull</option>}
                </select>
              </div>
            </div>
          </div>}


          {/* Add your modal content here */}


          {showModalError && (<p class="help is-danger">{pullError}</p>)}
          {peopleAlreadyInRoom.map((person, index) => (
            <div key={index} style={{ marginTop: '5px' }} className="field">
              <button className={`button is-danger ${loadingClearPerson[index] ? 'is-loading' : ''}`} onClick={() => {
                setLoadingClearPerson(loadingClearPerson.map((item, itemIndex) => itemIndex === index ? true : item));
                handleClearRoom(getRoomUUIDFromUserID(person), false, index);
              }}>Clear {getNameById(person)}'s existing room</button>
            </div>
          ))}

        </section>
        <footer className="modal-card-foot" style={{ display: 'flex', justifyContent: 'space-between' }}>
          <button className={`button is-primary ${loadingSubmit ? 'is-loading' : ''}`} onClick={handleSubmit}>Update room</button>
          <button className={`button is-danger ${loadingClearRoom ? 'is-loading' : ''}`} onClick={() => {
            setLoadingClearRoom(true);
            handleClearRoom(selectedRoomObject.roomUUID, true, -1);
          }}>Clear room</button>
        </footer>
      </div>
    </div>
  );
}

export default BumpModal;