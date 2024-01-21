import React from 'react';
import 'bulma/css/bulma.min.css';
import BumpModal from './BumpModal';
import { useState, useContext } from 'react';
import FloorGrid from './FloorGrid';
import Recommendations from './Recommendations';
import { MyContext } from './MyContext';

function App() {
  const {
    currPage,
    setCurrPage,
    gridData,
    setGridData,
    userMap,
    isModalOpen,
    // setUserMap,
    dormMapping,
    onlyShowBumpableRooms,
    setOnlyShowBumpableRooms,
    getNameById,
    // cellColors
    selectedID,
    setSelectedID,
    rooms,
  } = useContext(MyContext);

  const [showNotification, setShowNotification] = useState(false);



  const handleDeleteClick = () => {
    setShowNotification(false);
  };


  // const tabs = ["Atwood", "East", "Drinkward", "Linde", "North", "South", "Sontag", "West", "Case"]
  // const { drawNumbers } = useContext(MyContext);




  const [activeTab, setActiveTab] = useState('Atwood');

  const handleNameChange = (event) => {
    setSelectedID(event.target.value);
  };
  const handleTabClick = (tab) => {
    setActiveTab(tab);
  };

  function updateGridData(roomData) {
    setShowNotification(true);
    setGridData((prevGridData) => {
      const updatedGridData = prevGridData.map((dorm) => {
        const updatedFloors = dorm.floors.map((floor) => {
          return floor.map((room) => {
            if (room.roomNumber === roomData.roomNumber) {
              return { ...room, ...roomData };
            }
            return room;
          });
        });
        return { ...dorm, floors: updatedFloors };
      });
      return updatedGridData;
    });
  }



  function getDrawNumberAndYear(id) {
    // Find the drawNumber object with the given full name
    if (!userMap) {
      return "Loading...";
    }

    // console.log(userMap[id]);
    if (userMap[id].InDorm !== 0) {
      // has in dorm
      return `${userMap[id].Year.charAt(0).toUpperCase() + userMap[id].Year.slice(1)} ${userMap[id].DrawNumber} ${dormMapping[userMap[id].InDorm]}`;
    }
    return `${userMap[id].Year.charAt(0).toUpperCase() + userMap[id].Year.slice(1)} ${userMap[id].DrawNumber}`
  }

  // function getDrawNumberByName(name) {
  //   const foundItem = drawNumbers.find((item) => item.name === name);
  //   return foundItem ? foundItem.drawNumber : null;
  // }

  const getRoom = (name) => {
    // given a name, get the room and dorm that they are in 
    // handle case with no room yet
    if (!rooms) {
      return "";
    }
    if (rooms) {
      for (let room of rooms) {
        console.log(room.Occupants);
        console.log(selectedID);

        if (room.Occupants && room.Occupants.includes(selectedID)) {
          return `You are in ${room.DormName} ${room.RoomID}`;
        }
      }
      return "You are not in a room yet.";
    }

  }
  // fetch('/users/idmap')
  //         .then(res => {
  //             console.log(res);
  //             console.log("just printed");
  //             return res.json();  // Parse the response data as JSON
  //         })
  //         .then(data => {
  //             console.log(data);
  //             console.log("success");
  //             setUserMap(data);
  //         })
  //         .catch(err => {
  //             console.log(err);
  //         })




  return (
    <div>

      <nav class="navbar" role="navigation" aria-label="main navigation">
        <div class="navbar-brand">
          <a class="navbar-item" href="https://ibb.co/c3D21bJ"><img src="https://i.ibb.co/SyRVPQN/Screenshot-2023-12-26-at-10-14-31-PM.png" alt="Screenshot-2023-12-26-at-10-14-31-PM" border="0" /></a>

          <a role="button" class="navbar-burger" aria-label="menu" aria-expanded="false" data-target="navbarBasicExample">
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
          </a>
        </div>

        <div id="navbarBasicExample" class="navbar-menu">
          <div class="navbar-start">
            <a class="navbar-item" onClick={() => setCurrPage("Home")}>
              Home
            </a>

            <a class="navbar-item" onClick={() => setCurrPage("Recommendations")}>
              Recommendations
            </a>

            <div class="navbar-item has-dropdown is-hoverable">
              <a class="navbar-link">
                More
              </a>

              <div class="navbar-dropdown">
                <a class="navbar-item">
                  About
                </a>
                <a class="navbar-item">
                  Jobs
                </a>
                <a class="navbar-item">
                  Contact
                </a>

                <a class="navbar-item">
                  Report an issue
                </a>
              </div>
            </div>
          </div>

          <div class="navbar-end">
            <div class="navbar-item">
              <div class="buttons">
                <a class="button is-primary">
                  <strong>Sign up</strong>
                </a>
                <a class="button is-light">
                  Log in
                </a>
              </div>
            </div>
          </div>
        </div>
      </nav>
      {isModalOpen && <BumpModal />}
      {showNotification && (<div class="notification is-primary m-2">
        <button onClick={handleDeleteClick} class="delete "></button>
        Your room status has been updated. Please check that everything is still the way you'd like it to be!
      </div>)}
      <section class="section">
        <div style={{ textAlign: 'center' }}>

          <h1 className="title">Welcome back, <strong>{getNameById(selectedID)}</strong>. <br /> </h1>
          <h2 className="subtitle">You are <strong>{getDrawNumberAndYear(selectedID)}</strong>. {getRoom(selectedID)} <br />Click on any room you'd like to change!</h2>

          <div style={{ display: 'flex', justifyContent: 'center' }}>
            <select className="select" onChange={handleNameChange}>
              <option value="">This isn't me</option>
              {userMap && Object.keys(userMap)
                .sort((a, b) => {
                  const nameA = `${userMap[a].FirstName} ${userMap[a].LastName}`.toUpperCase();
                  const nameB = `${userMap[b].FirstName} ${userMap[b].LastName}`.toUpperCase();
                  if (nameA < nameB) {
                    return -1;
                  }
                  if (nameA > nameB) {
                    return 1;
                  }
                  return 0;
                })
                .map((key, index) => (
                  <option key={index} value={key}>
                    {userMap[key].FirstName} {userMap[key].LastName}
                  </option>
                ))}
            </select>
          </div>

        </div>

      </section>

      {currPage == "Home" && <section class="section">
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

            {gridData.map((dorm) => (
              <li
                key={dorm.dormName}
                className={activeTab === dorm.dormName ? 'is-active' : ''}
                onClick={() => handleTabClick(dorm.dormName)}
              >

                <a>{dorm.dormName}</a>
              </li>
            ))}
          </ul>
        </div>

        {/* Left column is room draw, right side is tips */}
        <div class="columns">

          <div class="column">
            {gridData.map((dorm) => (

              <div key={dorm.dormName} className={activeTab === dorm.dormName ? '' : 'is-hidden'}>

                {dorm.floors
                  .filter((floor) => floor.floorNumber % 2 === 0)
                  .sort((a, b) => Number(a.floorNumber) - Number(b.floorNumber))  // Convert to numbers before comparing
                  .map((floor, floorIndex) => (
                    <div style={{ paddingBottom: 20 }} className="container" key={floorIndex}>
                      <h2 class="subtitle has-text-centered">Floor {floor.floorNumber + 1}</h2>
                      <FloorGrid gridData={floor} />
                    </div>
                  ))}
              </div>
            ))}
          </div>
          <div class="column">
            {gridData.map((dorm) => (

              <div key={dorm.dormName} className={activeTab === dorm.dormName ? '' : 'is-hidden'}>

                {dorm.floors
                  .filter((floor) => floor.floorNumber % 2 !== 0)
                  .sort((a, b) => Number(a.floorNumber) - Number(b.floorNumber))  // Convert to numbers before comparing
                  .map((floor, floorIndex) => (
                    <div style={{ paddingBottom: 20 }} className="container" key={floorIndex}>
                      <h2 class="subtitle has-text-centered">Floor {floor.floorNumber + 1}</h2>
                      <FloorGrid gridData={floor} />
                    </div>
                  ))}
              </div>
            ))}
          </div>
          <div class="column">
            {gridData.map((dorm) => (

              <div key={dorm.dormName} className={activeTab === dorm.dormName ? '' : 'is-hidden'}>
                <p>{dorm.description}</p>
                {/* TODO!!! */}
                {/* {dorm.imageLinks.map((link, index) => (
                  <img src={link} alt="dorm" key={index} />
                ))} */}

              </div>
            ))}
          </div>
        </div>

        {/* {(activeTab === 'Recommendations' && (<Recommendations />))} */}

      </section>}
      {currPage == "Recommendations" && <section class="section">
        <Recommendations gridData={gridData} setCurrPage={setCurrPage} />
      </section>}


      <footer class="footer">
        <div class="content has-text-centered">
          <p>
            <strong>Digital Draw</strong> by Serena Mao & Tom Lam. Email smao@g.hmc.edu or tlam@g.hmc.edu with questions.
          </p>
        </div>
      </footer>


    </div>

  );
}


export default App;