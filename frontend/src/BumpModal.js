import { jwtDecode } from "jwt-decode";
import React, { useContext, useEffect, useState } from 'react';
import Select from "react-select";
import AdminBumpModalFunctions from './AdminBumpModalFunctions';
import { MyContext } from './MyContext';

// Utility function to simplify gender preferences display
const simplifyGenderPreferences = (preferences) =>
{
    if (!preferences || !preferences.length) return [];

    const simplifiedPrefs = [...preferences];
    const hasCisWoman = simplifiedPrefs.includes('Cis Woman');
    const hasTransWoman = simplifiedPrefs.includes('Trans Woman');
    const hasCisMan = simplifiedPrefs.includes('Cis Man');
    const hasTransMan = simplifiedPrefs.includes('Trans Man');

    // Replace Cis Woman and Trans Woman with Woman if both exist
    if (hasCisWoman && hasTransWoman)
    {
        // Remove both individual preferences
        const indexCisWoman = simplifiedPrefs.indexOf('Cis Woman');
        simplifiedPrefs.splice(indexCisWoman, 1);

        const indexTransWoman = simplifiedPrefs.indexOf('Trans Woman');
        simplifiedPrefs.splice(indexTransWoman, 1);

        // Add combined preference
        simplifiedPrefs.push('Woman');
    }

    // Replace Cis Man and Trans Man with Man if both exist
    if (hasCisMan && hasTransMan)
    {
        // Remove both individual preferences
        const indexCisMan = simplifiedPrefs.indexOf('Cis Man');
        simplifiedPrefs.splice(indexCisMan, 1);

        const indexTransMan = simplifiedPrefs.indexOf('Trans Man');
        simplifiedPrefs.splice(indexTransMan, 1);

        // Add combined preference
        simplifiedPrefs.push('Man');
    }

    return simplifiedPrefs;
};

// Helper function to check if a room is a Drinkward triple in suite side
const isDrinkwardSuiteTriple = (roomNumber, dormId) =>
{
    const drinkwardTripleInDormPullExceptions = ["123C", "124C", "221C", "222C", "223C", "224C", "321C", "322C", "323C", "324C"];
    return dormId === 8 || drinkwardTripleInDormPullExceptions.includes(roomNumber);
};

function BumpModal()
{
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
    // New state to store original values when focus changes
    const [originalOccupants, setOriginalOccupants] = useState({});
    // Add state for tracking clear room stats
    const [clearRoomStats, setClearRoomStats] = useState({
        clearRoomCount: 0,
        maxDailyClears: 10,
        remainingClears: 10,
        isBlacklisted: false
    });
    // Add state for room clearing confirmation
    const [showClearConfirmation, setShowClearConfirmation] = useState(false);
    const [clearConfirmationStep, setClearConfirmationStep] = useState(0);
    const [roomToClear, setRoomToClear] = useState(null);
    const [roomToClearCloseModal, setRoomToClearCloseModal] = useState(false);
    const [roomToClearPersonIndex, setRoomToClearPersonIndex] = useState(-1);

    useEffect(() =>
    {
        // If the selected suite or room changes, change the people who can pull 
        if (selectedSuiteObject)
        {
            console.log("selectedSuiteObject", selectedSuiteObject);
            console.log("selectedItem", selectedItem);
            let isDrinkwardTripleException = false;
            if (selectedSuiteObject.rooms[0])
            {
                console.log("selectedSuiteObject.rooms[0]", selectedSuiteObject.rooms[0]);
                isDrinkwardTripleException = isDrinkwardSuiteTriple(selectedItem, selectedSuiteObject.rooms[0].dorm);
                if (isDrinkwardTripleException)
                {
                    console.log("This is a drinkward triple in suite-side");
                }
            }
            const otherRooms = selectedSuiteObject.rooms;
            const otherOccupants = [];
            const otherRoomsWhoCanAlternatePull = [];
            console.log("otherRooms", otherRooms);
            for (let room of otherRooms)
            {
                if (room.roomNumber !== selectedItem && room.maxOccupancy === 1 && room.occupant1 !== 0 && room.pullPriority.pullType === 1)
                {
                    console.log("room", room);
                    otherOccupants.push([room.occupant1, room.roomUUID]);
                }
                if (selectedRoomObject.maxOccupancy === 2 && room.roomNumber !== selectedItem && room.maxOccupancy === 2)
                {
                    otherRoomsWhoCanAlternatePull.push([room.roomNumber, room.roomUUID]);
                }

            }
            console.log("otherOccupants", otherOccupants);
            setRoomsWhoCanAlternatePull(otherRoomsWhoCanAlternatePull);
            setPeopleWhoCanPullSingle(otherOccupants);
        }
    }, [selectedSuiteObject, selectedItem]);

    // Load clear room stats when component mounts
    useEffect(() =>
    {
        fetchClearRoomStats();
    }, [refreshKey]);

    // Function to fetch clear room stats
    const fetchClearRoomStats = () =>
    {
        fetch(`${process.env.REACT_APP_API_URL}/users/clear-room-stats`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('jwt')}`
            }
        })
            .then(response => response.json())
            .then(data =>
            {
                if (!data.error)
                {
                    setClearRoomStats(data);
                }
            })
            .catch(error =>
            {
                console.error("Error fetching clear room stats:", error);
            });
    };

    // Function to check for 403 blacklist responses
    const handleBlacklistCheck = (response) =>
    {
        if (response.status === 403)
        {
            return response.json().then(data =>
            {
                if (data.blacklisted)
                {
                    // Update the user's blacklist status
                    setClearRoomStats(prev => ({ ...prev, isBlacklisted: true }));
                    setPullError(data.error || 'Your account has been temporarily restricted due to excessive room clearing. Please contact an administrator.');
                    setShowModalError(true);
                    return true;
                }
                return false;
            });
        }
        return Promise.resolve(false);
    };

    const handlePullMethodChange = (e) =>
    {
        print(pullMethod);
        setPullMethod(e.target.value);
    };
    const closeModal = () =>
    {
        setLoadingSubmit(false);
        setLoadingClearPerson(loadingClearPerson.map((person) => false));
        setLoadingClearRoom(false);
        setShowModalError(false);
        setPullError("");
        setIsModalOpen(false);
    };
    const handleDropdownChange = (index, value) =>
    {
        print(value);
        const updatedselectedOccupants = [...selectedOccupants];
        updatedselectedOccupants[index - 1] = value;
        setSelectedOccupants(updatedselectedOccupants);
        print(selectedOccupants);
        setPeopleAlreadyInRoom([]);
        setShowModalError(false);
        setPullError("");

    };

    const handleSubmit = async (e) =>
    {  // Declare handleSubmit as async
        // Handle form submission logic here
        setLoadingSubmit(true);
        e.preventDefault();
        if (pullMethod.startsWith("Alt Pull"))
        {
            let roomUUID = pullMethod.slice("Alt Pull ".length).trim();
            if (await canIAlternatePull(roomUUID))
            {  // Wait for canIBePulled to complete
                print("This room was successfully alternative pulled");
                closeModal();
            } else
            {
                setLoadingSubmit(false);
                setShowModalError(true);
            }
        }
        else if (/^\d+$/.test(pullMethod))
        {
            // commented console.log ("Pull method is a number");
            // pullMethod only includes number, implying that you were pulled by someone else
            if (await canIBePulled())
            {  // Wait for canIBePulled to complete
                print("This room was successfully pulled by someone else in the suite");
                closeModal();
            } else
            {
                setLoadingSubmit(false);
                setShowModalError(true);
            }
        } else if (pullMethod === "Lock Pull")
        {
            // lock pulled 
            if (await canILockPull())
            {  // Wait for canIBePulled to complete
                print("This room was successfully lock pulled");
                closeModal();
            } else
            {
                setLoadingSubmit(false);
                setShowModalError(true);
            }
        } else if (pullMethod === "Alternate Pull")
        {
            // Pulled with 2nd best number of this suite
            // if (await canIAlternatePull()) {  // Wait for canIBePulled to complete
            //   print("This room was successfully pulled with 2nd best number of this suite");
            //   closeModal();
            // } else {
            //   setLoadingSubmit(false);
            //   setShowModalError(true);
            // }
        } else
        {
            // pullMethod is either Lock Pull or Pulled themselves 
            if (await canIBump())
            {  // Wait for canIBePulled to complete

                print("This room pulled themselves");
                closeModal();
                setRefreshKey(prev => prev + 1);
            } else
            {
                setLoadingSubmit(false);
                setShowModalError(true);
            }

        }


    };

    const performRoomAction = (pullType, pullLeaderRoom = null) =>
    {
        // commented console.log ("performing room actoin");
        setLoadingSubmit(true);
        return new Promise((resolve) =>
        {
            fetch(`${process.env.REACT_APP_API_URL}/rooms/${selectedRoomObject.roomUUID}`, {
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
                .then(response =>
                {
                    // Check for blacklist first
                    return handleBlacklistCheck(response).then(isBlacklisted =>
                    {
                        if (isBlacklisted)
                        {
                            setLoadingSubmit(false);
                            resolve(false);
                            return null;
                        }
                        return response.json().then(data =>
                        {
                            // Add status to the data object for error handling
                            return { ...data, status: response.status };
                        });
                    });
                })
                .then(data =>
                {
                    if (!data) return; // If blacklisted, skip this part

                    if (data.error)
                    {
                        if (handleErrorFromTokenExpiry(data))
                        {
                            return;
                        };
                        if (data.error === "One or more of the proposed occupants is already in a room")
                        {
                            // commented console.log ("Someone's already there rrror:");
                            // commented console.log (data.occupants);

                            // But wait: if these are the same people in the current room, handle clearing the room in the backend
                            // Check if all occupants are in the current room
                            if (data.occupants.length !== 0 && data.occupants.every(occupant => selectedRoomObject.roomUUID === getRoomUUIDFromUserID(occupant)))
                            {
                                // Clear the room and retry
                                console.log("That's the case!");
                                setLoadingSubmit(true);
                                handleClearRoom(selectedRoomObject.roomUUID, false, -1)
                                    .then(() => performRoomAction(pullType, pullLeaderRoom))
                                    .then((data) =>
                                    {
                                        // commented console.log (data);
                                        if (data === true)
                                        {
                                            setRefreshKey(prev => prev + 1);
                                            closeModal();
                                            // commented console.log ("THERE HAS NOT BEEN A ERROR");

                                        } else
                                        {
                                            setLoadingSubmit(false);
                                            setShowModalError(true);
                                            // commented console.log ("THERE HAS BEEN A ERROR");
                                            resolve(false);
                                        }
                                        resolve(true);
                                    });
                                return;
                            } else
                            {
                                setPeopleAlreadyInRoom((data.occupants));
                                setLoadingClearPerson(data.occupants.map((person) => false));

                                const names = data.occupants.map(getNameById).join(', ');
                                setPullError("Please remove " + names + " from their existing room");
                            }
                        }
                        else if (data.error === "One or more of the proposed occupants is not preplaced" && data.occupants)
                        {
                            // Handle the case for non-preplaced occupants
                            setPullError(data.error);
                            // You could also show which occupants aren't preplaced if needed
                            // const names = data.occupants.map(getNameById).join(', ');
                            // setPullError(`One or more of the proposed occupants is not preplaced: ${names}`);
                        }
                        else
                        {
                            setPullError(data.error);
                        }
                        setLoadingSubmit(false);
                        setShowModalError(true);
                        resolve(false);
                    } else
                    {
                        // commented console.log ("Refreshing and setting");
                        setRefreshKey(prev => prev + 1);
                        resolve(true);
                    }
                })
                .catch((error) =>
                {
                    console.error(error);
                    setLoadingSubmit(false);
                    setPullError("An unexpected error occurred. Please try again.");
                    setShowModalError(true);
                    setRefreshKey(prev => prev + 1);
                    resolve(false);
                });
        });
    }

    const canIBump = () => performRoomAction(1);
    const canIBePulled = () => performRoomAction(2, peopleWhoCanPullSingle.find(person => person[0] === Number(pullMethod))[1]);
    const canILockPull = () => performRoomAction(3);
    const canIAlternatePull = (roomUUID) =>
    {
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

    const handleClearRoom = (roomUUID, closeModalBool, personIndex) =>
    {
        return new Promise((resolve) =>
        {
            fetch(`${process.env.REACT_APP_API_URL}/rooms/clear/${roomUUID}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`
                }
            })
                .then(response =>
                {
                    // Check for blacklist first
                    return handleBlacklistCheck(response).then(isBlacklisted =>
                    {
                        if (isBlacklisted)
                        {
                            setLoadingClearRoom(false);
                            if (personIndex !== -1)
                            {
                                setLoadingClearPerson(loadingClearPerson.map((item, itemIndex) => itemIndex === personIndex ? false : item));
                            }
                            resolve(false);
                            return null;
                        }
                        return response.json();
                    });
                })
                .then(data =>
                {
                    if (!data) return; // If blacklisted, skip this part

                    if (data.error)
                    {
                        if (handleErrorFromTokenExpiry(data))
                        {
                            return;
                        };
                        setPullError(data.error);
                        setIsModalOpen(true);
                        setShowModalError(true);
                        resolve(false);

                        if (data.blacklisted)
                        {
                            // Update the user's blacklist status
                            setClearRoomStats(prev => ({ ...prev, isBlacklisted: true }));
                        }
                    } else
                    {
                        // Update clear room stats from response
                        if (data.clearRoomCount !== undefined)
                        {
                            setClearRoomStats({
                                clearRoomCount: data.clearRoomCount,
                                maxDailyClears: data.maxDailyClears,
                                remainingClears: data.remainingClears,
                                isBlacklisted: data.isBlacklisted || false
                            });
                        }

                        // no error 
                        // commented console.log ("Refreshing and settingloadClearPeron");
                        setRefreshKey(prev => prev + 1);
                        resolve(true);
                        if (personIndex !== -1)
                        {
                            setLoadingClearPerson(loadingClearPerson.filter((_, itemIndex) => itemIndex !== personIndex));
                            setPeopleAlreadyInRoom(peopleAlreadyInRoom.filter((_, itemIndex) => itemIndex !== personIndex));
                            setShowModalError("");
                        }
                        if (closeModalBool)
                        {
                            closeModal();
                        }
                    }
                })
                .catch((error) =>
                {
                    setRefreshKey(prev => prev + 1);
                    resolve(true);
                    if (closeModalBool)
                    {
                        closeModal();
                    } else
                    {
                        // On tap action for clearing other room
                        // Also clear the error and the button 
                        setShowModalError(false);
                        setPullError("");
                        setPeopleAlreadyInRoom([]);
                    }
                });
        });
    }

    const initiateRoomClear = (roomUUID, closeModalBool, personIndex) =>
    {
        // Store the room details for when confirmation is complete
        setRoomToClear(roomUUID);
        setRoomToClearCloseModal(closeModalBool);
        setRoomToClearPersonIndex(personIndex);

        // Show the first confirmation step
        setShowClearConfirmation(true);
        setClearConfirmationStep(1);
    };

    const confirmClearRoom = () =>
    {
        // Move to second confirmation step
        setClearConfirmationStep(2);
    };

    const cancelClearRoom = () =>
    {
        // Reset confirmation state
        setShowClearConfirmation(false);
        setClearConfirmationStep(0);
        setRoomToClear(null);
        setRoomToClearCloseModal(false);
        setRoomToClearPersonIndex(-1);

        // Reset loading state if needed
        setLoadingClearRoom(false);
        if (roomToClearPersonIndex !== -1)
        {
            setLoadingClearPerson(loadingClearPerson.map((item, itemIndex) =>
                itemIndex === roomToClearPersonIndex ? false : item
            ));
        }
    };

    const executeClearRoom = () =>
    {
        // Hide confirmation dialog
        setShowClearConfirmation(false);
        setClearConfirmationStep(0);

        // Set the appropriate loading state
        if (roomToClearPersonIndex === -1)
        {
            setLoadingClearRoom(true);
        } else
        {
            setLoadingClearPerson(loadingClearPerson.map((item, itemIndex) =>
                itemIndex === roomToClearPersonIndex ? true : item
            ));
        }

        // Execute the actual clear room operation
        handleClearRoom(roomToClear, roomToClearCloseModal, roomToClearPersonIndex);

        // Reset the stored values
        setRoomToClear(null);
        setRoomToClearCloseModal(false);
        setRoomToClearPersonIndex(-1);
    };

    // Function to format the minutes until reset
    const formatTimeUntilReset = (minutes) =>
    {
        if (minutes < 60)
        {
            return `${minutes} minutes`;
        } else
        {
            const hours = Math.floor(minutes / 60);
            const remainingMinutes = minutes % 60;
            if (remainingMinutes === 0)
            {
                return `${hours} hour${hours > 1 ? 's' : ''}`;
            } else
            {
                return `${hours} hour${hours > 1 ? 's' : ''} ${remainingMinutes} min`;
            }
        }
    };

    return (

        <div className="modal is-active">
            <div className="modal-background" onClick={closeModal}></div>
            <div className="modal-card">
                <header className="modal-card-head">
                    <p className="modal-card-title">
                        {selectedRoomObject.pullPriority.isPreplaced ? "Can't edit preplaced room" : `Edit Room ${selectedItem}`}
                    </p>
                    <button className="delete" aria-label="close" onClick={closeModal}></button>
                </header>
                <section className="modal-card-body">
                    {/* Display error message once at the top of the modal */}
                    {showModalError && (
                        <div className="notification is-danger" style={{ marginBottom: '15px' }}>
                            <p>{pullError}</p>
                        </div>
                    )}

                    {/* Add Gender Preferences display */}
                    {selectedSuiteObject && selectedSuiteObject.genderPreferences && selectedSuiteObject.genderPreferences.length > 0 && (
                        <div className="notification is-warning" style={{ marginBottom: '15px' }}>
                            <p className="has-text-weight-bold">This suite has the following gender preferences:</p>
                            <p>{simplifyGenderPreferences(selectedSuiteObject.genderPreferences).join(' or ')}</p>
                            <p className="mt-2">
                                <strong>Note:</strong> Knowingly violating these gender preferences is an <strong>Honor Code violation</strong>.
                            </p>
                        </div>
                    )}

                    <AdminBumpModalFunctions closeModal={closeModal} />

                    {!selectedRoomObject.pullPriority.isPreplaced && <div>
                        <div>
                            <label className="label">{`Reassign Occupant${selectedRoomObject.maxOccupancy > 1 ? "s" : ""}`}</label>

                            {[1, 2, 3, 4].slice(0, selectedRoomObject.maxOccupancy).map((index) => (
                                <div className="field" key={index}>
                                    <div className="control">
                                        <div style={{ marginBottom: "10px", width: 300 }}>
                                            <Select
                                                placeholder={`Search for occupant ${index}...`}
                                                value={
                                                    selectedOccupants[index - 1]
                                                }
                                                menuPortalTarget={document.body}
                                                onFocus={() =>
                                                {
                                                    // Store current value in originalOccupants state
                                                    setOriginalOccupants(prev => ({
                                                        ...prev,
                                                        [index - 1]: selectedOccupants[index - 1]
                                                    }));

                                                    // Clear the current value
                                                    const updatedOccupants = [...selectedOccupants];
                                                    updatedOccupants[index - 1] = null;
                                                    setSelectedOccupants(updatedOccupants);
                                                }}
                                                onBlur={() =>
                                                {
                                                    // If no new selection was made, restore the original value
                                                    if (!selectedOccupants[index - 1] && originalOccupants[index - 1])
                                                    {
                                                        const updatedOccupants = [...selectedOccupants];
                                                        updatedOccupants[index - 1] = originalOccupants[index - 1];
                                                        setSelectedOccupants(updatedOccupants);
                                                    }
                                                }}
                                                styles={{
                                                    menuPortal: base => ({ ...base, zIndex: 9999 }),
                                                    container: base => ({
                                                        ...base,
                                                        width: '100%',
                                                    }),
                                                    control: base => ({
                                                        ...base,
                                                        backgroundColor: 'var(--input-bg)',
                                                        borderColor: 'var(--border-color)',
                                                        color: 'var(--text-color)',
                                                        width: '100%',
                                                    }),
                                                    option: (provided, state) => ({
                                                        ...provided,
                                                        backgroundColor: state.isFocused
                                                            ? 'var(--primary-color)'
                                                            : document.body.classList.contains('dark-mode')
                                                                ? 'var(--card-bg)'
                                                                : provided.backgroundColor,
                                                        color: state.isFocused
                                                            ? '#ffffff'
                                                            : 'var(--text-color)',
                                                    }),
                                                    singleValue: (provided) => ({
                                                        ...provided,
                                                        color: 'var(--text-color)',
                                                    }),
                                                    menu: (provided) => ({
                                                        ...provided,
                                                        backgroundColor: 'var(--card-bg)',
                                                    }),
                                                    input: (provided) => ({
                                                        ...provided,
                                                        color: 'var(--text-color)',
                                                    }),
                                                }}
                                                onChange={(selectedOption) => handleDropdownChange(index, selectedOption)}
                                                options={userMap && Object.keys(userMap)
                                                    .sort((a, b) =>
                                                    {
                                                        const nameA = `${userMap[a].FirstName} ${userMap[a].LastName}`;
                                                        const nameB = `${userMap[b].FirstName} ${userMap[b].LastName}`;
                                                        return nameA.localeCompare(nameB);
                                                    })
                                                    .filter((key) => (((jwtDecode(credentials).email === "tlam@g.hmc.edu") || (jwtDecode(credentials).email === "smao@g.hmc.edu")) || Number(userMap[key].Year) !== 0)) // Replace 'YourCondition' with the condition you want to check
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
                                    { (selectedRoomObject.maxOccupancy === 1 || isDrinkwardSuiteTriple(selectedItem, selectedRoomObject.dorm)) && peopleWhoCanPullSingle.map((item, index) => (
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


                    {peopleAlreadyInRoom.map((person, index) => (
                        <div key={index} style={{ marginTop: '5px' }} className="field">
                            <button className={`button is-danger ${loadingClearPerson[index] ? 'is-loading' : ''}`} onClick={() =>
                            {
                                initiateRoomClear(getRoomUUIDFromUserID(person), false, index);
                            }}>Clear {getNameById(person)}'s existing room</button>
                        </div>
                    ))}

                    {/* Room Clear Confirmation Modal */}
                    {showClearConfirmation && (
                        <div className="modal is-active">
                            <div className="modal-background"></div>
                            <div className="modal-card">
                                <header className="modal-card-head">
                                    <p className="modal-card-title">
                                        {clearConfirmationStep === 1 ? "Warning: Room Clear Operation" : "Final Confirmation"}
                                    </p>
                                </header>
                                <section className="modal-card-body">
                                    {clearConfirmationStep === 1 ? (
                                        <div>
                                            <div className="notification is-warning">
                                                <p className="has-text-weight-bold">Warning: Clearing a room is a dangerous operation</p>
                                                <p>This will remove ALL occupants from the room and cannot be undone.</p>
                                                <p>Occupants will need to re-select their rooms.</p>
                                                <p className="mt-2">Do you want to proceed?</p>
                                            </div>
                                        </div>
                                    ) : (
                                        <div>
                                            <div className="notification is-danger">
                                                <p className="has-text-weight-bold">FINAL WARNING</p>
                                                <p>You are about to permanently clear this room. This action:</p>
                                                <ul style={{ marginLeft: '20px', marginTop: '10px', listStyleType: 'disc' }}>
                                                    <li>Cannot be undone</li>
                                                    <li>Will remove all occupants</li>
                                                    <li>May cause disruption to the room selection process</li>
                                                </ul>
                                                <p className="mt-3 has-text-weight-bold">Are you absolutely certain you want to continue?</p>
                                            </div>
                                        </div>
                                    )}
                                </section>
                                <footer className="modal-card-foot" style={{ justifyContent: 'space-between' }}>
                                    <button className="button" onClick={cancelClearRoom}>Cancel</button>
                                    {clearConfirmationStep === 1 ? (
                                        <button className="button is-warning" onClick={confirmClearRoom}>Proceed to Confirmation</button>
                                    ) : (
                                        <button className="button is-danger" onClick={executeClearRoom}>Yes, Clear Room</button>
                                    )}
                                </footer>
                            </div>
                        </div>
                    )}

                </section>
                <footer className="modal-card-foot" style={{ display: 'flex', justifyContent: 'space-between', flexDirection: 'column' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%', gap: '10px' }}>
                        <button className={`button is-primary ${loadingSubmit ? 'is-loading' : ''}`} style={{ flex: '1', maxWidth: '150px' }} onClick={handleSubmit}>Update room</button>
                        <button
                            className={`button is-danger ${loadingClearRoom ? 'is-loading' : ''}`}
                            style={{ flex: '1', maxWidth: '150px' }}
                            onClick={() =>
                            {
                                initiateRoomClear(selectedRoomObject.roomUUID, true, -1);
                            }}
                            disabled={clearRoomStats.remainingClears <= 0 || clearRoomStats.isBlacklisted || selectedRoomObject.pullPriority.isPreplaced}
                            title={selectedRoomObject.pullPriority.isPreplaced ? "Cannot clear a preplaced room" : ""}
                        >
                            Clear room
                        </button>
                    </div>
                    {clearRoomStats.isBlacklisted ? (
                        <div className="notification is-danger" style={{ marginTop: '10px', padding: '10px', fontSize: '0.85rem' }}>
                            Your account has been temporarily restricted due to excessive room clearing. Please contact an administrator.
                        </div>
                    ) : clearRoomStats.remainingClears <= 3 ? (
                        <div className="notification is-warning" style={{ marginTop: '10px', padding: '10px', fontSize: '0.85rem' }}>
                            Warning: You are approaching the daily limit for room clearing. Further clearing may result in account restriction.
                            {clearRoomStats.resetsInMinutes && (
                                <div style={{ marginTop: '5px' }}>
                                    Limits reset at midnight Pacific Time which is ({formatTimeUntilReset(clearRoomStats.resetsInMinutes)})
                                </div>
                            )}
                        </div>
                    ) : (
                        <div className="notification is-info" style={{ marginTop: '10px', padding: '10px', fontSize: '0.85rem' }}>
                            Clearing rooms should be done sparingly. Excessive usage will result in account restriction.
                            {clearRoomStats.resetsInMinutes && (
                                <div style={{ marginTop: '5px' }}>
                                    Limits reset at midnight Pacific Time which ({formatTimeUntilReset(clearRoomStats.resetsInMinutes)})
                                </div>
                            )}
                        </div>
                    )}
                </footer>
            </div>
        </div>
    );
}

export default BumpModal;