/*################################################
    グローバル変数等宣言
##################################################*/
//キャンバスの準備
let can = document.getElementById("can");
let con = can.getContext("2d");
const CANVAS_W = 1000;
const CANVAS_H = 600;
can.width = CANVAS_W
can.height = CANVAS_H;
can.style.border = "4px solid"; 

//カーソルが移動できる一番左、右の牌(配列の一番左、右)を定義
// const MOSTLEFT = 0;
// const MOSTRIGHT = 13;

//カーソルが選択している牌配列の要素番号
// let cursorSelecting = 0;

// //捨て牌を選ぶ際の疑似待ち時間のループ
// let isWaitingSelection = false;

// let playerTiles = [];
// let cpuTiles = [[],[],[]];
// // let deck = [];
// let player_tsumo_tile = [];
// let player_throw_away_tile = [];
// let cpu_tsumo_tile = [];
// let cpu_throw_away_tile = [[],[],[]];
// let playerThrownAwayTiles = [[],[],[],[]];
// let finish = false;
// //let is_win = false;

// //王牌の数
// const THE_NUMBER_OF_TILES_IN_WAN_PAI = 14;
// // //数牌の数字(一から)
// // let number_tiles_number;
// // //数牌の数字(九まで)
// // const MAX_NUM = 9;
// // //同じ牌の個数
// // const SAME_TILES = 4;
// //ひとつの色の数牌の個数
// const THE_NUMBER_OF_NUMBER_TILES_IN_ONE_COLOR = 36;
// //手牌の初期個数
// // const THE_NUMBER_OF_TILES_IN_HAND = 13;
// // //数牌のマンズ、ピンズ、ソーズの色を設定する
// // const NUMBER_TILES_COLORS = ["m","p","s"];
// // //理牌する時用に牌の並びの優先度も決めておく
// // //あとで優先度をキーとして配列を参照するので0からの方が都合が良い
// // let  tiles_priNum;

// // //字牌に色は無いのでz(字)で統一
// // const CHARACTER_TILES_COLOR = "z";
// //字牌の東南西北白發中を設定する
// const CHARACTER_TILES_TYPES = 
//   ["t","n","sh","p","hk","hts","ch"];

// //牌の個数(可変)
// let the_number_of_tiles_in_hand;

// //自家手牌配列の中での理牌時の優先度キーの要素番号
// const priNum = 2;
// //取得対象画像の取得先チップ画像内での座標
// let tile_x_inImg;
// let tile_x_inImg;
// //取得先のチップ画像内での牌の画像の大きさを取得する
// const TILE_W = tiles[0].w;
// const TILE_H = tiles[0].h;

// //CPUの数
// const CPUS = 3;

// //牌を描画する際の描画先の初期座標
// let draw_x;
// let draw_y;

// //捨てる牌を選ぶ待ち時間のループを定義
// let interval_id;
// let when_select_tile = false;
// const LOOP_PACE = 100;

// //牌を選択するカーソルの大きさ・場所を定義
// const cursor_w = 5;
// const cursor_h = 15;
// let cursor_x;
// let cursor_y;

// //カーソルが移動できる一番左、右の牌(配列の一番左、右)を定義
// const MOSTLEFT = 0;
// let MOSTRIGHT;

// //カーソルが選択している牌配列の要素番号
// let cursorSelecting = 0;

// //捨て牌を選ぶ際の疑似待ち時間のループ
// let isWaitingSelection = false;

// //CPUの行動遅延のループ
// let is_wait = false;
// let cpu_n = 0;

// //エレベーターのサイズ
// const ELEVATOR_S = 150;

// //河の一行に置ける牌の数
// const TILES_IN_RIVER = 6;

// //牌を取得した大きさから縮小するための値
// const shrink = 0.7;

// //他家が牌を捨てる時の遅延
// const WAIT_TIME = 300;

/*#################################################
    処理
###################################################*/
//画像などすべてのロードを終えてからプログラム開始
window.onload = function(){
  main();
}


/*#################################################
    関数
###################################################*/

//メインの関数
async function main(){
  let [deck, playerTiles, cpuTiles] = prepareGame();

  drawAll(deck, playerTiles, cpuTiles);

  let playerThrownAwayTiles;
  [deck, playerTiles, playerThrownAwayTiles] = await playerAction(deck, playerTiles);

  let cpuThrownAwayTiles = [[], [], []];
  for(let i=0; len<CPUS; i++){
    [deck, cpuTiles, cpuThrownAwayTiles[i]] = await cpusAction(deck, cpuTiles[i]);
  }
}

/*---------------------------------------------------
  ゲームの準備関連
----------------------------------------------------*/
//ゲームの全準備。自家の手牌、他家の手牌、山を設定する
// TODO 関数にするまでもない？=> 見やすそうだし分けよう
function prepareGame(){
  let deck = initDeck();
  let playerTiles;
  let cpuTiles;

  [deck, playerTiles, cpuTiles] = preparePlayersTiles(deck);
  return [deck, playerTiles, cpuTiles];
}


//デッキの準備。全ての牌を作る
function initDeck(){
  /*概要：山の情報の設定
    数牌：[牌の色、牌の数字、優先度],[〃, 〃, 〃],...
    字牌：[牌の色(字)、牌の種類、優先度],[〃, 〃, 〃],...
    これらを配列deckに入れ、シャッフルして返す*/

  let deck = [];
  let num = 1;
  let key = 0;
  const COLORS = ["m","p","s"];
  const Z_TYPES = ["t","n","sh","p","hk","hts","ch"];
  const tilesInOneColor = 9;
  const SAME_TILES = 4;

  //数牌の牌を作る
  //各色ごとに一から九の牌を4つずつ作る
  COLORS.forEach(function(color){
    for(let i=0; i<tilesInOneColor; i++){
      for(let j=0; j<SAME_TILES; j++){
        //山の配列に1牌ごとに[色、数、優先度]の情報を入れる
        deck.push({color, num, key});
      }
      num++;
      key++;
    }
    num = 1;
  });

  const COLOR = "z";
  //字牌の牌を作る
  //字牌の各種類ごとに同じ牌を4つずつ作る
  Z_TYPES.forEach(function(type){
    for(let i=0; i< SAME_TILES; i++){
        deck.push({color:COLOR, type, key});
      }
      num++;
      key++;
  });
  //牌をシャッフルして返す
  arrayShuffle(deck);
  return deck;
}


//牌をシャッフルする
function arrayShuffle(array){
  for(let i = (array.length - 1); 0 < i; i--){
    // 0〜(i+1)の範囲で値を取得
    let r = Math.floor(Math.random() * (i + 1));

    // 要素の並び替えを実行
    let tmp = array[i];
    array[i] = array[r];
    array[r] = tmp;
  }
}


//自家・他家の手牌を準備
function preparePlayersTiles(preDeck){
  /*概要：自家と他家の手牌情報の設定
    山から13枚引く*/
  let deck = [...preDeck];
  const TILES_IN_HAND = 13;

  //自家が山から13枚引く
  let playerTiles = [];
  for(let i=0; i<TILES_IN_HAND; i++){
    playerTiles.push(deck.pop());
  }
  riiPai(playerTiles);

  let cpuTiles = [[],[],[]];

  //他家が山から13枚引く×3回
  cpuTiles.forEach((tiles) => {
    for(let i=0; i<TILES_IN_HAND; i++){
      tiles.push(deck.pop());
    }
    riiPai(tiles);
  });

  return [deck, playerTiles, cpuTiles];
}

//理牌する
function riiPai(tiles){
  //手牌の優先度キーを参照してバブルソートで昇順に並び替える
  for(let i=0, len=tiles.length; i<len; i++){
    for(let j=tiles.length-1; j>i; j--){
      if(tiles[j].key < tiles[j-1].key){
        tmp = tiles[j];
        tiles[j] = tiles[j-1];
        tiles[j-1] = tmp;
      }
    }
  }
}


/*---------------------------------------------------
  自家のアクション関連
----------------------------------------------------*/
//自家の全てのアクション
async function playerAction(preDeck, prePlayersTiles){
  const WAN_PAI = 14;

  //残り牌数が王牌以上だったら行動する
  if(preDeck.length > WAN_PAI){
    let deck = [...preDeck];
    let playerTiles = [...prePlayersTiles];

    //ツモる
    [deck, playerTiles] = getTile(preDeck, prePlayersTiles);

    //打牌
    let playerThrownAwayTiles;
    [playerTiles, playerThrownAwayTiles[playerThrownAwayTiles.length]] = await throwAway(deck, playerTiles);

    return [deck, playerTiles, playerThrownAwayTiles]
  }
  else{
    // finish = true;
    console.log("山が無くなりました！");
  }
}


function getTile(deck, preTiles) {
  let newTiles = [...preTiles, deck.pop()];
  return [deck, newTiles];
}


// //自家の打牌アクション
async function throwAway(deck, playerTiles){
  let cursorSelecting = 0;
  const MOST_LEFT = 0;
  const MOST_RIGHT = 14
  let timeoutId;
  let isWait_f = true;

  document.addEventListener("keydown", (e) => {
    //自家が捨て牌を選ぶ時間じゃない時＝自家の捨て牌選択ループ外では動かない
    switch(e.keyCode){
      case 37:  //左
        isWaiting(isWait_f, () => {
          //カーソルが一番左じゃない限り左にカーソルを移動させる
          if(cursorSelecting != MOST_LEFT ){
            cursorSelecting--;
            //新しいカーソルの位置を描画し直す
            drawAll(deck, playerTiles, cpuTiles);
            drawCursor();
          }}
        );
      break;
      case 39:  //右
        isWaiting(isWait_f, () => {
           if(cursorSelecting != MOST_RIGHT ){
            cursorSelecting++;
            drawAll(deck, playerTiles, cpuTiles);
            drawCursor();
          }
        })
        break;
      //エンターを押したら捨て牌選択待ちのループを解除して次の処理に行けるようにする
      //つまり、エンターで捨て牌を決定して捨てるということ
      case 13:  //エンター
        isWaiting(isWait_f, () => {
          clearTimeout(timeoutId);
          isWait_f = false;
        })

      // case 82:
      //   if(finish){
      //     clear_game();
      //     main();
      //   }
  }});


  let cancel;
  (async ()=>{
    const a = asyncSetTimeout(1000,throwAwayFunc)
    cancel = a.cancel
    await a.exec() // ここで設定した時間分処理を待ったあとasyncFuncを実行する
    // doSomething()
  })()

  cancel();
}


function asyncSetTimeout(msec, func){
  let timeoutId
  let r
  const exec = function(){
    return new Promise((res) => {
      r = res
      timeoutId = setTimeout(async () => {
          timeoutId = null;
          await func();
          res();
      },msec);
    });
  }
  return {
    exec,
    cancel: () => {
        if (timeoutId) {
          clearTimeout(timeoutId)
          timeoutId = null
          r()
        }
    }
  }
}



function asyncTimeout(msec, func){
  let timeoutId;
  let r;

  const exec = () => new Promise((res) => {
    r = res;
    //打牌の時間切れの処理
    timeoutId = setTimeout(() => {
      //手牌から適当に選ぶ
      let rand = Math.floor(Math.random() * playerTiles.length);
      let playerThrownAwayTiles = playerTiles[rand];

      //適当に選んだ牌を捨てる
      playerTiles.splice(rand, 1);
      riiPai(playerTiles);

      resolve([playerTiles, playerThrownAwayTiles]);
    }, 5000)
  })
}




//プレイヤーの打牌選択待ち状態か判定
function isWaiting(isWait_f, fn){
  if (isWait_f){
    fn();
  }
}


// //自家がテンパイしているか判定する
// function judge_tempai(){
//   //ToDo テンパイ判定(もシャンテン数判定も)プログラムが難しすぎて挫折
//   //次の処理
//   is_wait = true;
//   wait_cpu_loop();
// }


/*---------------------------------------------------
  他家のアクション関連
----------------------------------------------------*/
//他家がツモって捨てる(ツモ切り)
function cpuAction(i, deck, cpuTiles){
  //ツモできる牌がある(残り牌が王牌の数より多い)時だけ動く
  if(deck.length > THE_NUMBER_OF_TILES_IN_WAN_PAI){
    //ツモ
    getTile(deck, cpuTiles[i]);

    //ToDo とりあえず手牌の中から配列番号をランダムに選んで捨ててる
    let rand = Math.floor(Math.random() * cpuTiles.length);
    let cpuThrownAwayTiles = cpuTiles[i][rand]
    //cpu_throw_away_tile[i].push(cpuTiles[i][rand]);

    playerThrownAwayTiles[i+1].push(cpuTiles[i][rand]);
    cpuTiles[i].splice(rand, 1);

    //捨ててから理牌
    //してもそんなに意味ないけど
    riiPai(cpuTiles[i]);

    //描画し直し
    drawAll();

    if(cpu_n == 2){
      cpu_n = 0;
      playerAction();
    }
    else{
      cpu_n++;
      is_wait = true;
      wait_cpu_loop();
    }
  }
  else{
    console.log("山が無くなりました！")
    finish = true;
  }
}


// // //
// // function wait_cpu_loop(){
// //   if(is_wait){
// //     is_wait = false;
// //     setTimeout(wait_cpu_loop, WAIT_TIME);
// //   }
// //   else{
// //     cpu_actions(cpu_n);
// //   }
// // }



// // /*---------------------------------------------------
// //   描画関連
// // ----------------------------------------------------*/
//全てを描画
function drawAll(deck, playerTiles, cpuTiles){
  clearCanvas();
  drawTiles(playerTiles);
  drawCpuTiles(cpuTiles);
  drawElevator();
  drawTilesCount(deck);
  // drawThrownAwayTiles();
}

//手牌を描画
function drawTiles(playerTiles){
  //手牌の数を宣言
  let tilesInHand = playerTiles.length;

  //牌を描画する初期値を設定する
  const TILE_W = tiles[0].w;
  const TILE_H = tiles[0].h;
  let draw_x = (can.width / 2) - (TILE_W * (tilesInHand / 2));
  let draw_y = can.height - TILE_H;

  let tile_x_inImg;
  let tile_y_inImg;

  //手牌13枚の牌配列の優先度キーから画像を取得し描画する
  for(let i=0; i < tilesInHand; i++){
    //自家手牌の優先度キーから、対象の牌の画像が取得先のチップ画像内でどの座標にあるのかを得る
    tile_x_inImg = tiles[playerTiles[i].key].x_in_img;
    tile_y_inImg = tiles[playerTiles[i].key].y_in_img;

    //描画する
    con.drawImage( tiles_img, tile_x_inImg, tile_y_inImg, TILE_W, TILE_H,
      draw_x, draw_y, TILE_W, TILE_H);

    //次の描画位置を設定してあげる
    draw_x += TILE_W;
  }
  //カーソル描画の際にこれらの座標を使うのでバグが起こらないよう初期値に戻しておく
  draw_x = (can.width / 2) - (TILE_W * ( tilesInHand / 2));
  draw_y = can.height - TILE_H;
}


//捨て牌を選ぶカーソルを描画
function drawCursor(){
  const TILE_W = 47;
  const TILE_H = 63;
  const CURSOR_W = 3;
  const CURSOR_H = 5;
  let space = 3;
  let tilesInHand = 14;
  let draw_x = (can.width / 2) - (TILE_W * (tilesInHand / 2));
  let draw_y = can.height - TILE_H;
  //カーソルの位置を指定
  let cursor_x = draw_x + ((TILE_W - CURSOR_W)/2) + (cursorSelecting * TILE_W);
  let cursor_y = draw_y - (CURSOR_H + space);

  //とりあえず緑の背景で目立つ黄色で描画
  con.fillStyle = "yellow";
  con.fillRect(cursor_x, cursor_y, cursor_w, cursor_h);
}

//キャンバスをクリアする
function clearCanvas(){
  con.clearRect(0, 0, can.width, can.height);
}

// //CPUの手牌を表示する(伏せているので実際の手牌は反映していない)
function drawCpuTiles(cpuTiles){
  let tilesInHand = cpuTiles[0].length;
  //cpuの手牌画像の取得先での座標と、画像の大きさを宣言
  //縦向きか横向きかで取得すべき座標と大きさが違うので配列で指定
  //それぞれ下家→対面→上家の情報を格納している
  const TILE_X_IN_IMG = [0,44,0];
  const TILE_Y_IN_IMG = [0, 0, 0];
  const TILE_W = [44, 33, 44];
  const TILE_H = [49, 59, 49];
  //上家と下家の手牌描画時の重なりを考慮した値
  const TILE_H_ADJUSTED = 32;

  //描画の初期位置を設定。各場所の中心に手牌が来るように設定
  let draw_x = [can.width-TILE_W[0],
          (can.width/2) - (TILE_W[1] * (tilesInHand/2)),
          0];
  let draw_y = [(can.height/2) - (TILE_H_ADJUSTED * (tilesInHand/2)),
          0,
          (can.height/2) - (TILE_H_ADJUSTED * (tilesInHand/2))];

  const CPUS = 3;
  //3人分手牌を設定した位置に描画
  //人数分、手牌の数分描画する
  for(let i=0;i<CPUS;i++){
    //手牌の長さを再宣言
    tilesInHand = cpuTiles[i].length;

    for(let j=0;j<tilesInHand;j++){
      //描画
      con.drawImage(tiles_in_hand_img, TILE_X_IN_IMG[i], TILE_Y_IN_IMG[i], TILE_W[i], TILE_H[i],
            draw_x[i], draw_y[i], TILE_W[i], TILE_H[i]);

      //次の描画位置を設定
      if(i==0 || i==2){
        draw_y[i] += TILE_H_ADJUSTED;
      }
      else{
        draw_x[i] += TILE_W[i];
      }
    }
  }
}

//エレベーターを描画する
function drawElevator(){
  const ELEVATOR_S = 120;
  con.fillStyle = "black";
  con.fillRect( can.width/2 - (ELEVATOR_S/2), can.height/2 - (ELEVATOR_S/2), ELEVATOR_S, ELEVATOR_S);
}

//山の残り枚数を表示する
function drawTilesCount(deck){
  let text = "残り：" + deck.length + "枚";
  let text_w = 120;
  con.font = "20px Arial";
  con.fillStyle = "white";
  con.fillText(text, can.width/2 - text_w/2, can.height/2);
}


// //どうにか分かりやすくしようとしたが挫折

// //捨てられた牌を河に描画する
// function drawThrownAwayTiles(){
//   //捨て牌が無かったら返す
//   if(playerThrownAwayTiles === undefined){
//     return;
//   }

//   //取得したい画像の取得先での座標と大きさを宣言しておく
//   let tile_x_inImg = [];
//   let tile_y_inImg = [];
//   //取得先の画像サイズ
//   //牌の向きによって取得したい画像が違うので大きさを配列に格納しておく
//   //順に自家→下家→対面→上家の取得画像の大きさを宣言している
//   let TILE_W_INIMG = [47, 63, 47, 63];
//   let TILE_H_INIMG = [63, 47, 63, 47];

//   //描画する画像サイズ
//   //取得先の画像サイズに調整する変数をかける
//   let TILE_W = [];
//   let TILE_H = [];

//   //取得した画像をキャンバスに表示できるように縮小するための変数
//   let shrink = 0.7;

//   const PLAYERS = 4;

//   for(let i=0; i<PLAYERS; i++){
//     TILE_W[i] = Math.floor(TILE_W_INIMG[i] * shrink);
//     TILE_H[i] = Math.floor(TILE_H_INIMG[i] * shrink);
//   }

//   //描画位置の初期設定
//   //エレベーターと河の隙間
//   const SPACE = 25;
//   const draw_x = [(can.width/2) - (TILE_W[0] * (TILES_IN_RIVER/2)),
//           (can.width/2) + ELEVATOR_S/2 + SPACE,
//           (can.width/2) + (TILE_W[2] * (TILES_IN_RIVER/2)) - TILE_W[2],
//           (can.width/2) - ELEVATOR_S/2 - TILE_W[3] - SPACE];

//   let draw_y = [(can.height/2) + (ELEVATOR_S/2) + SPACE,
//           (can.height/2) + TILE_H[1] * (TILES_IN_RIVER/2) - TILE_H[1],
//           (can.height/2) - (ELEVATOR_S/2) - TILE_H[2] - SPACE,
//           (can.height/2) - TILE_H[3] * (TILES_IN_RIVER/2)];

//   let river_row;

//   //人数分、捨て牌の数の分描画する
//   for(let i=0; i<CPUS + 1; i++){
//     //捨て牌が無かったら終わる
//     if(!playerThrownAwayTiles[i].length){
//       break;
//     }

//     river_row = 0;

//     for(let j=0; j<playerThrownAwayTiles[i].length; j++){
//       //捨て牌の画像の取得先の座標を、捨て牌配列内の牌の優先度キーから取得する
//       tile_x_inImg[i] = tiles[playerThrownAwayTiles[i][j][priNum]].col_in_img * TILE_W_INIMG[i];
//       //優先度キーと牌の種類の関係は崩したくないが、同じ種類で向きが違う牌が出てくるとその関係維持が難しい
//       //なので苦肉の策として末尾の(i*TILE_H[0]*4)(←縦のサイズ63であれば[0]じゃなくても2でもいい)で優先度キーと牌の繋がりはそのままに取得する画像の座標を調整した
//       //ToDo Tilesクラスで定数this.TILE_W/hを使ってthis.x/y_in_imgを宣言してしまっていて大きさが違う画像の座標が指定しにくくなっているので直す
//       tile_y_inImg[i] = tiles[playerThrownAwayTiles[i][j][priNum]].row_in_img * TILE_H_INIMG[i] + (i * TILE_H_INIMG[0] * 4);

//       if(i == 0){
//         //もし6牌以上捨て牌がある場合はその分改行して描画する
//         if(Math.floor(j / TILES_IN_RIVER) - river_row){
//           draw_y[i] += TILE_H[i];
//           draw_x[i] = (can.width/2) - (TILE_W[0] * (TILES_IN_RIVER/2));
//           river_row++;
//         }

//         //描画
//         con.drawImage(tiles_img, tile_x_inImg[i], tile_y_inImg[i], TILE_W_INIMG[i], TILE_H_INIMG[i],
//           draw_x[i], draw_y[i], TILE_W[i], TILE_H[i]);
        
//         //次の描画位置を設定
//         draw_x[i] += TILE_W[i];
//       }
//       else if(i == 1){
//         //もし6牌以上捨て牌がある場合はその分改行して描画する
//         if(Math.floor(j / TILES_IN_RIVER) - river_row){
//           draw_x[i] += TILE_W[i];
//           draw_y[i] = (can.height/2) + TILE_H[1] * (TILES_IN_RIVER/2) - TILE_H[1];
//           river_row++;
//         }

//         //描画
//         con.drawImage(tiles_img, tile_x_inImg[i], tile_y_inImg[i], TILE_W_INIMG[i], TILE_H_INIMG[i],
//           draw_x[i], draw_y[i], TILE_W[i], TILE_H[i]);

//         //次の描画位置を設定
//         draw_y[i] -= TILE_H[i];
//       }
//       else if(i==2){
//         //もし6牌以上捨て牌がある場合はその分改行して描画する
//         if(Math.floor(j / TILES_IN_RIVER) - river_row){
//           draw_y[i] -= TILE_H[i];
//           draw_x[i] = (can.width/2) + (TILE_W[2] * (TILES_IN_RIVER/2)) - TILE_W[2];
//           river_row++;
//         }

//         //描画
//         con.drawImage(tiles_img, tile_x_inImg[i], tile_y_inImg[i], TILE_W_INIMG[i], TILE_H_INIMG[i],
//           draw_x[i], draw_y[i], TILE_W[i], TILE_H[i]);

//         //次の描画位置を設定
//         draw_x[i] -= TILE_W[i];      
//       }
//       else{
//         //もし6牌以上捨て牌がある場合はその分改行して描画する
//         if(Math.floor(j / TILES_IN_RIVER) - river_row){
//           draw_x[i] -= TILE_W[i];
//           draw_y[i] = (can.height/2) - TILE_H[3] * (TILES_IN_RIVER/2);
//           river_row++;
//         }

//         //描画
//         con.drawImage(tiles_img, tile_x_inImg[i], tile_y_inImg[i], TILE_W_INIMG[i], TILE_H_INIMG[i],
//           draw_x[i], draw_y[i], TILE_W[i], TILE_H[i]);

//         //次の描画位置を設定
//         draw_y[i] += TILE_H[i];
//       }
//     } 
//   }

//   function aiue(){

//   }
// }

// function addNewLine(fn){
//   if(Math.floor(j / TILES_IN_RIVER) - river_row){
//     fn();
//   }
// }

// /*
// function draw_a_tile_in_river(position, j, size1, size2){
//   //河に6個並んだら改行する
//   position += Math.floor(j/TILES_IN_RIVER) * size1 * shrink;
//   //描画
//   con.drawImage(tiles_img, tile_x_inImg[i], tile_x_inImg[i], TILE_W[i], TILE_H[i],
//     draw_x[i], draw_y[i], Math.floor(TILE_W[i]*0.7), Math.floor(TILE_H[i]*0.7));

//   console.log(tile_x_inImg, tile_x_inImg, TILE_W[i], TILE_H[i],
//     draw_x[i], draw_y[i], Math.floor(TILE_W[i]*0.7), Math.floor(TILE_H[i]*0.7));

//   //次の描画位置を設定
//   draw_x[i] += TILE_W[i] * shrink;
// }*/






// //リーチの表示をする
// function riich(){

// }

// //リーチ中、他家が捨てた牌がロンか判定する
// function judge_ron(){

// }

// //リーチ中、ツモる
// function player_in_riich_tusmo(){

// }

// //ツモった牌がアガリ牌か判定する
// function judge_tsumo(){
  
// }