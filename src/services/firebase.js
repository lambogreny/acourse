import firebase from 'firebase'
import { Observable, BehaviorSubject } from 'rxjs'

export default {
  init () {
    firebase.initializeApp({
      apiKey: 'AIzaSyBa6jPVcSPbLzjuei9-d0u9q7M-9rGjmb8',
      authDomain: 'acourse-d9d0a.firebaseapp.com',
      databaseURL: 'https://acourse-d9d0a.firebaseio.com',
      storageBucket: 'acourse-d9d0a.appspot.com',
      messagingSenderId: '582047384847'
    })
    this.currentUser = new BehaviorSubject()
    firebase.auth().onAuthStateChanged((user) => {
      this.currentUser.next(user)
    })
  },
  signInWithEmailAndPassword (email, password) {
    return Observable.fromPromise(firebase.auth().signInWithEmailAndPassword(email, password))
  },
  signInWithFacebook () {
    return Observable.fromPromise(firebase.auth().signInWithPopup(new firebase.auth.FacebookAuthProvider()))
  },
  createUserWithEmailAndPassword (email, password) {
    return Observable.fromPromise(firebase.auth().createUserWithEmailAndPassword(email, password))
  },
  sendPasswordResetEmail (email) {
    return Observable.fromPromise(firebase.auth().sendPasswordResetEmail(email))
  },
  signOut () {
    return Observable.fromPromise(firebase.auth().signOut())
  },
  onValue (path) {
    const ref = firebase.database().ref(path)
    return Observable.bindCallback(ref.on.bind(ref))('value')
      .map((snapshot) => snapshot.val())
  },
  onArrayValue (path) {
    const ref = firebase.database().ref(path)
    return Observable.bindCallback(ref.on.bind(ref))('value')
      .map((snapshot) => {
        const result = []
        snapshot.forEach((x) => {
          const v = x.val()
          v.id = x.key
          result.push(v)
        })
        return result
      })
  },
  upload (path, file) {
    const ref = firebase.storage().ref(path)
    return Observable.fromPromise(ref.put(file))
  },
  set (path, data) {
    const ref = firebase.database().ref(path)
    return Observable.fromPromise(ref.set(data))
  },
  push (path, data) {
    const ref = firebase.database().ref(path)
    return Observable.fromPromise(ref.push(data))
  },
  get timestamp () {
    return firebase.database.ServerValue.TIMESTAMP
  }
}
