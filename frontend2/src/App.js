import React, { useEffect, useState, useContext } from 'react';
import 'bulma/css/bulma.min.css';
import BumpModal from './BumpModal';
import FloorGrid from './FloorGrid';
import Recommendations from './Recommendations';
import { MyContext } from './MyContext';
import { GoogleLogin } from '@react-oauth/google';
import { jwtDecode } from "jwt-decode";
import SuiteNoteModal from './SuiteNoteModal';
import { googleLogout } from '@react-oauth/google';
import BumpFroshModal from './BumpFroshModal';
import Select from 'react-select';
import SettingsModal from './SettingsModal';


function App() {
  const options = [
    'one', 'two', 'three'
  ];
  const defaultOption = options[0];

  const {
    currPage,
    setCurrPage,
    gridData,
    userMap,
    isModalOpen,
    dormMapping,
    onlyShowBumpableRooms,
    setOnlyShowBumpableRooms,
    getNameById,
    selectedID,
    setSelectedID,
    rooms,
    isSuiteNoteModalOpen,
    credentials,
    setCredentials,
    lastRefreshedTime,
    activeTab,
    setActiveTab,
    isFroshModalOpen,
    getRoomUUIDFromUserID,
    roomRefs,
    setRefreshKey,
    handleErrorFromTokenExpiry,
    isSettingsModalOpen,
    setIsSettingsModalOpen,
    showFloorplans,
    setShowFloorplans,

  } = useContext(MyContext);

  const [showNotification, setShowNotification] = useState(false);

  useEffect(() => {
    const hasClosedNotification = localStorage.getItem('hasClosedNotification');
    setShowNotification(hasClosedNotification !== 'true');
  }, []);

  const handleCloseNotification = () => {
    // Store the notification state in local storage when the user closes the notification
    localStorage.setItem('hasClosedNotification', 'true');
    setShowNotification(false);
  };

  const [myRoom, setMyRoom] = useState("You are not in a room yet."); // to show what room current user is in
  const [isBurgerClicked, setIsBurgerClicked] = useState(false);
  const [isInDorm, setIsInDorm] = useState(true);

  useEffect(() => {

    const thisRoom = getRoomObjectFromUserID(selectedID);
    if (thisRoom) {
      // commented console.log (thisRoom);
      // commented console.log (thisRoom.PullPriority);
      var pullPriority = thisRoom.PullPriority;
      if (pullPriority.inherited.valid) {
        pullPriority = pullPriority.inherited;
      }
      if (pullPriority.hasInDorm === true) {
        setIsInDorm(false);
      } else {
        setIsInDorm(true);
      }

    }

  }, [selectedID, rooms]);

  useEffect(() => {
    const storedCredentials = localStorage.getItem('jwt');
    if (storedCredentials) {
      // commented console.log ("use effect");
      // commented console.log (storedCredentials);
      // commented console.log ("end use efect");
      setCredentials(storedCredentials);
    }
  }, []);

  // useEffect(() => {
  //   // Check for stored credentials on component mount
  //   const storedCredentials = localStorage.getItem('jwt');
  //   if (storedCredentials) {
  //     // If credentials exist, decode and set them
  //     setCredentials(storedCredentials);
  //   }
  // }, []);

  const handleSuccess = (credentialResponse) => {
    // decode the credential

    const decoded = jwtDecode(credentialResponse.credential);
    // commented console.log (decoded);

    setCredentials(credentialResponse.credential);
    localStorage.setItem('jwt', credentialResponse.credential); // originally stored credentialREsponse
    // localStorage.setItem('jwt', credentials); // originally stored credentialREsponse
    // commented console.log (typeof credentialResponse.credential);
  };

  const handleError = () => {
    // commented console.log ('Login Failed');
    // Optionally, handle login failure (e.g., by clearing stored credentials)
  };

  const handleLogout = () => {
    setCredentials(null);
    localStorage.removeItem('jwt');
    googleLogout();
  };

  const getRoomObjectFromUserID = (userID) => {
    if (rooms) {
      for (let room of rooms) {

        if (room.Occupants && room.Occupants.includes(Number(selectedID))) {


          return room;
        }
      }


    }
    return null;
  }

  const canUserToggleInDorm = (userID) => {
    userID = Number(userID);
    const usersRoom = getRoomObjectFromUserID(userID);
    if (!userMap) {
      return -1;
    }

    if (!usersRoom) {
      if (dormMapping[userMap[userID].InDorm]) {
        return 0;
      }
      return -1;
    }
    if (userMap[userID].InDorm === 0) {
      return -1;
    }
    if (dormMapping[userMap[userID].InDorm] === usersRoom.DormName) {
      return 1;
    } else if (dormMapping[userMap[userID].InDorm]) {
      return 0;
    }
    return -1;

  }



  useEffect(() => {
    // updates room that thei current user is in every time the selected user or the room data changes
    if (!rooms || !Array.isArray(rooms)) {
      return;
    }
    if (rooms) {
      for (let room of rooms) {

        if (room.Occupants && room.Occupants.includes(Number(selectedID))) {


          setMyRoom(`${room.DormName} ${room.RoomID}`);
          return;
        }
      }
      setMyRoom(`no room yet`);


    }
  }, [selectedID, rooms]);



  // Save state to localStorage whenever it changes
  useEffect(() => {
    localStorage.setItem('activeTab', activeTab);
  }, [activeTab]);


  const handleNameChange = (newID) => {
    setSelectedID(newID);
  };

  const handleTabClick = (tab) => {
    setActiveTab(tab);
  };

  function getDrawNumberAndYear(id) {
    // Find the drawNumber in laymans terms with the given id, including in-dorm status
    // ex: given 2, returns Sophomore 46
    if (!userMap) {
      return "Loading...";
    } else if (userMap[id].InDorm && userMap[id].InDorm !== 0) {
      // has in dorm
      return `${userMap[id].Year.charAt(0).toUpperCase() + userMap[id].Year.slice(1)} ${userMap[id].DrawNumber} with ${dormMapping[userMap[id].InDorm]} In-Dorm`;
    }
    return `${userMap[id].Year.charAt(0).toUpperCase() + userMap[id].Year.slice(1)} ${userMap[id].DrawNumber}`
  }



  // Component for each floor, to show even and odd floors separately
  const FloorDisplay = ({ gridData, filterCondition }) => {
    const filteredFloors = gridData.flatMap(dorm => dorm.floors.filter(floor => filterCondition(floor.floorNumber)));

    return (
      <div className="column">
        <div style={showFloorplans ? { width: '50vw' } : {}}>
          {gridData.map((dorm) => (
            <div key={dorm.dormName} className={activeTab === dorm.dormName ? '' : 'is-hidden'}>
              {dorm.floors
                .filter((floor) => filterCondition(floor.floorNumber))
                .sort((a, b) => Number(a.floorNumber) - Number(b.floorNumber))  // Convert to numbers before comparing
                .map((floor, floorIndex) => (
                  <div style={{ paddingBottom: 20 }} className="container" key={floorIndex}>
                    <h2 className="subtitle has-text-centered">Floor {floor.floorNumber + 1}</h2>
                    {floor.floorName && <p className="subtitle has-text-centered">{floor.floorName}</p>}
                    <FloorGrid gridData={floor} />
                  </div>
                ))}
            </div>
          ))}
        </div>
      </div>
    );
  };
  const handleTakeMeThere = (myLocationString) => {
    const words = myLocationString.split(' ');
    if (words.length === 2) {
      setActiveTab(words[0]);
    }

    // Assume `selectedID` is the ID of the selected room
    const roomUUID = getRoomUUIDFromUserID(selectedID); // Replace this with the actual function to get the room UUID

    // Delay the scrolling until after the tab has finished switching
    setTimeout(() => {
      const roomRef = roomRefs.current[roomUUID];
      if (roomRef) {
        roomRef.scrollIntoView({ behavior: 'smooth' });
      }
    }, 0);
  }

  const handleForfeit = () => {
    if (localStorage.getItem('jwt')) {
      fetch(`https://www.cs.hmc.edu/~tlam/digitaldraw/api/rooms/indorm/${getRoomUUIDFromUserID(selectedID)}`, {
          method: 'POST',
          headers: {
              'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
          },
      })
        .then(res => {
          return res.json();
        })
        .then(data => {
          setRefreshKey(prevKey => prevKey + 1);
          if (handleErrorFromTokenExpiry(data)) {
            return;
          };
          const thisRoom = getRoomObjectFromUserID(selectedID);
          // commented console.log ("BRUHMOMENT");
          // commented console.log (thisRoom);
          // commented console.log (thisRoom.PullPriority.hasInDorm);
          setIsInDorm(prev => !prev);

        })
        .catch(err => {
          // commented console.log (err);
        })


    }
  }

  return (
    <div>
      <nav class="navbar" role="navigation" aria-label="main navigation">
        <div class="navbar-brand">
          <a class="navbar-item" href="#"><img src="https://i.ibb.co/SyRVPQN/Screenshot-2023-12-26-at-10-14-31-PM.png" alt="Screenshot-2023-12-26-at-10-14-31-PM" border="0" /></a>

          {/* <a role="button" class="navbar-burger" aria-label="menu" aria-expanded="false" data-target="navbarBasicExample" onClick={() => setIsBurgerClicked(true)}>
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
          </a> */}
          {(!credentials && window.innerWidth < 1024) &&
            <GoogleLogin auto_select={true}
              onSuccess={handleSuccess}
              onError={handleError}
            />}
          {(credentials && window.innerWidth < 1024) && <a class="button is-danger" onClick={handleLogout}>
            <strong>Log Out</strong>
          </a>}
        </div>



        <div id="navbarBasicExample" class="navbar-menu">
          <div class="navbar-start">
            <div class="navbar-item">
              <h2 >Last refresh: {lastRefreshedTime.toLocaleTimeString()}</h2>

            </div>
          </div>

          <div className="navbar-end">
            <div className="navbar-item">
              <div className="buttons">
                {credentials &&
                  <div>

                    <a className="button is-secondary">

                      <strong>Welcome, {jwtDecode(credentials).given_name} </strong>
                    </a>
                    <button className="button is-primary" onClick={() => setIsSettingsModalOpen(prev => !prev)}>
                      Settings
                    </button>
                    <a className="button is-danger" onClick={handleLogout}>

                      <strong>Log Out</strong>
                    </a>
                  </div>

                }

                {!credentials &&
                  <GoogleLogin auto_select={true}
                    onSuccess={handleSuccess}
                    onError={handleError}
                  />}

              </div>
            </div>
          </div>
        </div>
      </nav>
      {isModalOpen && <BumpModal />}
      {isSuiteNoteModalOpen && <SuiteNoteModal />}
      {isFroshModalOpen && <BumpFroshModal />}
      {isSettingsModalOpen && <SettingsModal />}

      {(credentials && showNotification) && <div class="notification is-link" style={{ marginLeft: '20px', marginRight: '20px' }}>
        <button class="delete" onClick={handleCloseNotification}></button>
        UI Update: 1) Preplaced rooms now appear slightly darker by default. 2) Unbumpable rooms are now always black.
      </div>}




      {!credentials && <section class="section">
        <div style={{ textAlign: 'center' }}>
          <h1 className="title">Welcome to Digital Draw!</h1>
          <h2 className="subtitle">Please log in with your HMC email to continue.</h2>
        </div>
      </section>}
      {(credentials) && <section class="section">
        <div style={{ textAlign: 'center' }}>
          <h1 className="title">Welcome to DigiDraw!</h1>
          <div style={{ display: 'flex', flexDirection: 'row', alignItems: 'center', justifyContent: "center", flexWrap: 'wrap' }}>

            <Select
              placeholder={`Select a user`}
              value={userMap &&
              {
                value: selectedID,
                label: `${userMap[selectedID].FirstName} ${userMap[selectedID].LastName}`
              }
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
              onChange={(selectedOption) => handleNameChange(selectedOption.value)}
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
            <div style={{ marginLeft: "10px" }}>
              <h3 className="subtitle is-5">
                is <strong>{getDrawNumberAndYear(selectedID)}</strong>,
                in <strong>
                  {myRoom !== `no room yet` ?
                    <a href="#" onClick={() => handleTakeMeThere(myRoom)} style={{ textDecoration: 'underline' }}>{myRoom}</a>
                    : 'no room yet'}
                </strong>

              </h3>


            </div>


          </div>
          {canUserToggleInDorm(selectedID) !== -1 &&
            <div>
              <input
                type="checkbox"
                id="toggleInDorm"
                name="toggleInDorm"
                disabled={canUserToggleInDorm(selectedID) === 0}
                checked={isInDorm}
                onChange={handleForfeit}
              />
              {canUserToggleInDorm(selectedID) === 1 &&
                <label htmlFor="toggleInDorm" style={{ marginLeft: '5px' }}>Forfeit In-Dorm for their current single</label>
              }
              {canUserToggleInDorm(selectedID) === 0 &&
                <label htmlFor="toggleInDorm" style={{ marginLeft: '5px' }}>To forfeit their in-dorm, pull into a single.</label>
              }
            </div>
          }
        </div>
      </section>}

      {(credentials && currPage === "Home") && <section class="section">
        <label className="checkbox" style={{ display: 'flex', alignItems: 'center' }}>
          <input
            type="checkbox"
            checked={onlyShowBumpableRooms}
            onChange={() => setOnlyShowBumpableRooms(!onlyShowBumpableRooms)}
          />
          <span style={{ marginLeft: '0.5rem' }}>Darken rooms I can't pull</span>
        </label>


        <div className="tabs is-centered">
          <ul>

            {gridData.length === 9 && gridData.map((dorm) => (
              (
                <li
                  key={dorm.dormName}
                  className={activeTab === dorm.dormName ? 'is-active' : ''}
                  onClick={() => handleTabClick(dorm.dormName)}
                >
                  <a>{dorm.dormName}</a>
                </li>
              )
            ))}
          </ul>
        </div>

        {userMap && <div class="columns">
          {!showFloorplans && gridData
            .filter(dorm => dorm.dormName === activeTab)
            .flatMap(dorm => dorm.floors)
            .map((_, floorIndex) => (
              activeTab === "Case" ? (
                floorIndex < 2 && <FloorDisplay gridData={gridData} filterCondition={(floorNumber) => floorNumber === floorIndex} />
              ) : (
                <FloorDisplay gridData={gridData} filterCondition={(floorNumber) => floorNumber === floorIndex} />
              )
            ))}
          {showFloorplans && (
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              {gridData
                .filter(dorm => dorm.dormName === activeTab)
                .flatMap(dorm => dorm.floors)
                .map((_, floorIndex) => (
                  <div style={{ display: 'flex', alignItems: 'center' }}>
                    <FloorDisplay gridData={gridData} filterCondition={(floorNumber) => floorNumber === floorIndex} />
                    <img
                      src={`https://www.cs.hmc.edu/~tlam/digitaldraw/Floorplans/floorplans_${activeTab.toLowerCase()}_${floorIndex + 1}.png`}
                      alt={`Floorplan for floor ${floorIndex}`}
                      style={{ maxWidth: '40vw' }} // Add this line
                    />
                  </div>
                ))}
            </div>
          )}
        </div>}


      </section>}

      {currPage === "Recommendations" && <section class="section">
        {/* <Recommendations gridData={gridData} setCurrPage={setCurrPage} /> */}
      </section>}


      <footer class="footer">
        <div class="content has-text-centered">
          <p>
            <strong>Digital Draw</strong> by Serena Mao & Tom Lam. Email smao@g.hmc.edu, or tlam@g.hmc.edu with questions.
          </p>
        </div>
      </footer>


    </div>

  );
}


export default App;