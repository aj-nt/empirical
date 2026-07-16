<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                xmlns:math="http://exslt.org/math"
                xmlns:vedic="urn:empirical:vedic"
                version="1.0">

  <!-- ── Vedic planets (9 grahas) ──────────────────────────────────────── -->
  <!-- 7 classical + Rahu (TrueNode). Ketu computed as Rahu + 180°. -->
  <xsl:variable name="vedicPlanets" select="',Sun,Moon,Mars,Mercury,Jupiter,Venus,Saturn,TrueNode,'"/>

  <!-- ── Sign names ────────────────────────────────────────────────────── -->
  <xsl:variable name="signs" select="'Aries,Taurus,Gemini,Cancer,Leo,Virgo,Libra,Scorpio,Sagittarius,Capricorn,Aquarius,Pisces'"/>

  <!-- ── Nakshatra names (27) ──────────────────────────────────────────── -->
  <xsl:variable name="nakshatras" select="'Ashwini,Bharani,Krittika,Rohini,Mrigashira,Ardra,Punarvasu,Pushya,Ashlesha,Magha,Purva Phalguni,Uttara Phalguni,Hasta,Chitra,Swati,Vishakha,Anuradha,Jyeshtha,Mula,Purva Ashadha,Uttara Ashadha,Shravana,Dhanishtha,Shatabhisha,Purva Bhadrapada,Uttara Bhadrapada,Revati'"/>

  <!-- ── Nakshatra lords (Vimshottari order, 9 repeating) ─────────────── -->
  <xsl:variable name="nakLords" select="'Ketu,Venus,Sun,Moon,Mars,Rahu,Jupiter,Saturn,Mercury'"/>

  <!-- ── Vimshottari dasha years per lord ──────────────────────────────── -->
  <xsl:variable name="dashaYears" select="'7,20,6,10,7,18,16,19,17'"/>

  <!-- ── Tithi names (30) ──────────────────────────────────────────────── -->
  <xsl:variable name="tithis" select="'Pratipada,Dwitiya,Tritiya,Chaturthi,Panchami,Shashthi,Saptami,Ashtami,Navami,Dashami,Ekadashi,Dwadashi,Trayodashi,Chaturdashi,Purnima,Pratipada,Dwitiya,Tritiya,Chaturthi,Panchami,Shashthi,Saptami,Ashtami,Navami,Dashami,Ekadashi,Dwadashi,Trayodashi,Chaturdashi,Amavasya'"/>

  <!-- ══════════════════════════════════════════════════════════════════════
       ROOT TEMPLATE
       ════════════════════════════════════════════════════════════════════ -->
  <xsl:template match="/BaseChart">
    <vedic:Chart xmlns:vedic="urn:empirical:vedic">
      <vedic:Name><xsl:value-of select="Identity/Name"/></vedic:Name>
      <vedic:Ayanamsa><xsl:value-of select="format-number(Positions/Ayanamsa, '0.00')"/></vedic:Ayanamsa>

      <!-- Planet signs (sidereal) -->
      <vedic:PlanetSigns>
        <xsl:for-each select="Positions/Planet[contains($vedicPlanets, concat(',', @name, ','))]">
          <xsl:sort select="@id" data-type="number"/>
          <vedic:PlanetSign>
            <vedic:Planet>
              <xsl:choose>
                <xsl:when test="@name = 'TrueNode'">Rahu</xsl:when>
                <xsl:otherwise><xsl:value-of select="@name"/></xsl:otherwise>
              </xsl:choose>
            </vedic:Planet>
            <vedic:Sign>
              <xsl:call-template name="lonToSign">
                <xsl:with-param name="lon" select="Sidereal/Lon"/>
              </xsl:call-template>
            </vedic:Sign>
            <vedic:Degree>
              <xsl:value-of select="format-number(Sidereal/Lon mod 30, '0.00')"/>
            </vedic:Degree>
            <vedic:Retrograde>
              <xsl:value-of select="Sidereal/Speed &lt; 0"/>
            </vedic:Retrograde>
          </vedic:PlanetSign>
        </xsl:for-each>
        <!-- Ketu (computed from Rahu) -->
        <xsl:call-template name="ketuSign"/>
      </vedic:PlanetSigns>

      <!-- Whole-sign houses (sidereal) -->
      <vedic:PlanetHouses>
        <xsl:for-each select="Positions/Planet[contains($vedicPlanets, concat(',', @name, ','))]">
          <xsl:sort select="@id" data-type="number"/>
          <vedic:PlanetHouse>
            <vedic:Planet>
              <xsl:choose>
                <xsl:when test="@name = 'TrueNode'">Rahu</xsl:when>
                <xsl:otherwise><xsl:value-of select="@name"/></xsl:otherwise>
              </xsl:choose>
            </vedic:Planet>
            <vedic:House>
              <xsl:call-template name="lonToWholeSignHouse">
                <xsl:with-param name="lon" select="Sidereal/Lon"/>
              </xsl:call-template>
            </vedic:House>
          </vedic:PlanetHouse>
        </xsl:for-each>
        <!-- Ketu -->
        <xsl:call-template name="ketuHouse"/>
      </vedic:PlanetHouses>

      <!-- Nakshatras -->
      <vedic:Nakshatras>
        <!-- ASC nakshatra -->
        <vedic:AscNakshatra>
          <xsl:call-template name="computeNakshatra">
            <xsl:with-param name="lon" select="Angles/ASC"/>
            <xsl:with-param name="label" select="'ASC'"/>
          </xsl:call-template>
        </vedic:AscNakshatra>
        <!-- Planet nakshatras -->
        <xsl:for-each select="Positions/Planet[contains($vedicPlanets, concat(',', @name, ','))]">
          <xsl:sort select="@id" data-type="number"/>
          <vedic:PlanetNakshatra>
            <vedic:Planet>
              <xsl:choose>
                <xsl:when test="@name = 'TrueNode'">Rahu</xsl:when>
                <xsl:otherwise><xsl:value-of select="@name"/></xsl:otherwise>
              </xsl:choose>
            </vedic:Planet>
            <xsl:call-template name="computeNakshatra">
              <xsl:with-param name="lon" select="Sidereal/Lon"/>
            </xsl:call-template>
          </vedic:PlanetNakshatra>
        </xsl:for-each>
        <!-- Ketu nakshatra -->
        <xsl:call-template name="ketuNakshatra"/>
      </vedic:Nakshatras>

      <!-- Vedic dignities (swakshetra/uchcha/neecha/peregrine) -->
      <vedic:Dignities>
        <xsl:for-each select="Positions/Planet[contains($vedicPlanets, concat(',', @name, ','))]">
          <xsl:sort select="@id" data-type="number"/>
          <xsl:call-template name="assessVedicDignity">
            <xsl:with-param name="planet" select="."/>
          </xsl:call-template>
        </xsl:for-each>
        <!-- Ketu dignity -->
        <xsl:call-template name="ketuDignity"/>
      </vedic:Dignities>

      <!-- Tithi -->
      <vedic:Tithi>
        <xsl:call-template name="computeTithi"/>
      </vedic:Tithi>

      <!-- Vimshottari dasha -->
      <vedic:Vimshottari>
        <xsl:call-template name="computeVimshottari"/>
      </vedic:Vimshottari>

      <!-- Nodes (Rahu/Ketu) -->
      <vedic:Nodes>
        <xsl:call-template name="computeNodes"/>
      </vedic:Nodes>

      <!-- Angles (sidereal) -->
      <vedic:Angles>
        <vedic:ASC>
          <xsl:call-template name="lonToSign">
            <xsl:with-param name="lon" select="Angles/ASC"/>
          </xsl:call-template>
          <xsl:text> </xsl:text>
          <xsl:value-of select="format-number(Angles/ASC mod 30, '0.00')"/>
        </vedic:ASC>
        <vedic:MC>
          <xsl:call-template name="lonToSign">
            <xsl:with-param name="lon" select="Angles/MC"/>
          </xsl:call-template>
          <xsl:text> </xsl:text>
          <xsl:value-of select="format-number(Angles/MC mod 30, '0.00')"/>
        </vedic:MC>
        <vedic:DSC>
          <xsl:call-template name="lonToSign">
            <xsl:with-param name="lon" select="Angles/DSC"/>
          </xsl:call-template>
          <xsl:text> </xsl:text>
          <xsl:value-of select="format-number(Angles/DSC mod 30, '0.00')"/>
        </vedic:DSC>
        <vedic:IC>
          <xsl:call-template name="lonToSign">
            <xsl:with-param name="lon" select="Angles/IC"/>
          </xsl:call-template>
          <xsl:text> </xsl:text>
          <xsl:value-of select="format-number(Angles/IC mod 30, '0.00')"/>
        </vedic:IC>
      </vedic:Angles>

    </vedic:Chart>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       KETU HELPERS (computed as Rahu + 180°)
       ════════════════════════════════════════════════════════════════════ -->

  <xsl:template name="ketuLon">
    <xsl:param name="rahuLon" select="Positions/Planet[@name='TrueNode']/Sidereal/Lon"/>
    <xsl:value-of select="($rahuLon + 180) mod 360"/>
  </xsl:template>

  <xsl:template name="ketuSign">
    <xsl:variable name="klon">
      <xsl:call-template name="ketuLon"/>
    </xsl:variable>
    <vedic:PlanetSign>
      <vedic:Planet>Ketu</vedic:Planet>
      <vedic:Sign>
        <xsl:call-template name="lonToSign">
          <xsl:with-param name="lon" select="$klon"/>
        </xsl:call-template>
      </vedic:Sign>
      <vedic:Degree>
        <xsl:value-of select="format-number($klon mod 30, '0.00')"/>
      </vedic:Degree>
      <vedic:Retrograde>true</vedic:Retrograde>
    </vedic:PlanetSign>
  </xsl:template>

  <xsl:template name="ketuHouse">
    <xsl:variable name="klon">
      <xsl:call-template name="ketuLon"/>
    </xsl:variable>
    <vedic:PlanetHouse>
      <vedic:Planet>Ketu</vedic:Planet>
      <vedic:House>
        <xsl:call-template name="lonToWholeSignHouse">
          <xsl:with-param name="lon" select="$klon"/>
        </xsl:call-template>
      </vedic:House>
    </vedic:PlanetHouse>
  </xsl:template>

  <xsl:template name="ketuNakshatra">
    <xsl:variable name="klon">
      <xsl:call-template name="ketuLon"/>
    </xsl:variable>
    <vedic:PlanetNakshatra>
      <vedic:Planet>Ketu</vedic:Planet>
      <xsl:call-template name="computeNakshatra">
        <xsl:with-param name="lon" select="$klon"/>
      </xsl:call-template>
    </vedic:PlanetNakshatra>
  </xsl:template>

  <xsl:template name="ketuDignity">
    <xsl:variable name="klon">
      <xsl:call-template name="ketuLon"/>
    </xsl:variable>
    <xsl:variable name="signIdx" select="floor($klon div 30)"/>
    <vedic:Dignity>
      <vedic:Planet>Ketu</vedic:Planet>
      <vedic:Sign>
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$signs"/>
          <xsl:with-param name="n" select="$signIdx + 1"/>
        </xsl:call-template>
      </vedic:Sign>
      <vedic:Swakshetra>
        <xsl:call-template name="checkVedicDomicile">
          <xsl:with-param name="planet" select="'Ketu'"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </vedic:Swakshetra>
      <vedic:Uchcha>
        <xsl:call-template name="checkVedicExaltation">
          <xsl:with-param name="planet" select="'Ketu'"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </vedic:Uchcha>
      <vedic:Neecha>
        <xsl:call-template name="checkVedicDebilitation">
          <xsl:with-param name="planet" select="'Ketu'"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </vedic:Neecha>
      <vedic:State>
        <xsl:call-template name="determineVedicState">
          <xsl:with-param name="planet" select="'Ketu'"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </vedic:State>
    </vedic:Dignity>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       CORE HELPERS
       ════════════════════════════════════════════════════════════════════ -->

  <!-- ── Longitude to sign name ────────────────────────────────────────── -->
  <xsl:template name="lonToSign">
    <xsl:param name="lon"/>
    <xsl:variable name="idx" select="floor($lon div 30)"/>
    <xsl:call-template name="nthToken">
      <xsl:with-param name="list" select="$signs"/>
      <xsl:with-param name="n" select="$idx + 1"/>
    </xsl:call-template>
  </xsl:template>

  <!-- ── nth comma-separated token ─────────────────────────────────────── -->
  <xsl:template name="nthToken">
    <xsl:param name="list"/>
    <xsl:param name="n"/>
    <xsl:choose>
      <xsl:when test="$n = 1">
        <xsl:value-of select="substring-before(concat($list, ','), ',')"/>
      </xsl:when>
      <xsl:otherwise>
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="substring-after($list, ',')"/>
          <xsl:with-param name="n" select="$n - 1"/>
        </xsl:call-template>
      </xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Sidereal whole-sign house ─────────────────────────────────────── -->
  <xsl:template name="lonToWholeSignHouse">
    <xsl:param name="lon"/>
    <xsl:variable name="asc" select="/BaseChart/Angles/ASC"/>
    <xsl:variable name="ascSignStart" select="floor($asc div 30) * 30"/>
    <xsl:variable name="house" select="(($lon - $ascSignStart + 360) mod 360) div 30 + 1"/>
    <xsl:value-of select="floor($house)"/>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       NAKSHATRA COMPUTATION
       ════════════════════════════════════════════════════════════════════ -->

  <xsl:template name="computeNakshatra">
    <xsl:param name="lon"/>
    <xsl:param name="label"/>  <!-- optional: 'ASC' for ascendant -->
    <xsl:variable name="nakIdx" select="floor($lon div 13.333333)"/>
    <xsl:variable name="nakDeg" select="$lon - ($nakIdx * 13.333333)"/>
    <xsl:variable name="pada" select="floor($nakDeg div 3.333333) + 1"/>
    <xsl:variable name="lordIdx" select="$nakIdx mod 9"/>

    <vedic:Nakshatra>
      <xsl:call-template name="nthToken">
        <xsl:with-param name="list" select="$nakshatras"/>
        <xsl:with-param name="n" select="$nakIdx + 1"/>
      </xsl:call-template>
    </vedic:Nakshatra>
    <vedic:Pada><xsl:value-of select="$pada"/></vedic:Pada>
    <vedic:Lord>
      <xsl:call-template name="nthToken">
        <xsl:with-param name="list" select="$nakLords"/>
        <xsl:with-param name="n" select="$lordIdx + 1"/>
      </xsl:call-template>
    </vedic:Lord>
    <vedic:NakDeg><xsl:value-of select="format-number($nakDeg, '0.00')"/></vedic:NakDeg>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       VEDIC DIGNITIES
       ════════════════════════════════════════════════════════════════════ -->

  <xsl:template name="assessVedicDignity">
    <xsl:param name="planet"/>
    <xsl:variable name="name">
      <xsl:choose>
        <xsl:when test="$planet/@name = 'TrueNode'">Rahu</xsl:when>
        <xsl:otherwise><xsl:value-of select="$planet/@name"/></xsl:otherwise>
      </xsl:choose>
    </xsl:variable>
    <xsl:variable name="lon" select="$planet/Sidereal/Lon"/>
    <xsl:variable name="signIdx" select="floor($lon div 30)"/>

    <vedic:Dignity>
      <vedic:Planet><xsl:value-of select="$name"/></vedic:Planet>
      <vedic:Sign>
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="$signs"/>
          <xsl:with-param name="n" select="$signIdx + 1"/>
        </xsl:call-template>
      </vedic:Sign>
      <vedic:Swakshetra>
        <xsl:call-template name="checkVedicDomicile">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </vedic:Swakshetra>
      <vedic:Uchcha>
        <xsl:call-template name="checkVedicExaltation">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </vedic:Uchcha>
      <vedic:Neecha>
        <xsl:call-template name="checkVedicDebilitation">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </vedic:Neecha>
      <vedic:State>
        <xsl:call-template name="determineVedicState">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </vedic:State>
    </vedic:Dignity>
  </xsl:template>

  <!-- ── Vedic domicile (swakshetra) ───────────────────────────────────── -->
  <xsl:template name="checkVedicDomicile">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:choose>
      <xsl:when test="$planet = 'Sun' and $signIdx = 4">true</xsl:when>
      <xsl:when test="$planet = 'Moon' and $signIdx = 3">true</xsl:when>
      <xsl:when test="$planet = 'Mars' and ($signIdx = 0 or $signIdx = 7)">true</xsl:when>
      <xsl:when test="$planet = 'Mercury' and ($signIdx = 2 or $signIdx = 5)">true</xsl:when>
      <xsl:when test="$planet = 'Jupiter' and ($signIdx = 8 or $signIdx = 11)">true</xsl:when>
      <xsl:when test="$planet = 'Venus' and ($signIdx = 1 or $signIdx = 6)">true</xsl:when>
      <xsl:when test="$planet = 'Saturn' and ($signIdx = 9 or $signIdx = 10)">true</xsl:when>
      <xsl:when test="$planet = 'Rahu' and $signIdx = 10">true</xsl:when>   <!-- Aquarius -->
      <xsl:when test="$planet = 'Ketu' and $signIdx = 7">true</xsl:when>    <!-- Scorpio -->
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Vedic exaltation (uchcha) ─────────────────────────────────────── -->
  <xsl:template name="checkVedicExaltation">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:choose>
      <xsl:when test="$planet = 'Sun' and $signIdx = 0">true</xsl:when>       <!-- Aries -->
      <xsl:when test="$planet = 'Moon' and $signIdx = 1">true</xsl:when>      <!-- Taurus -->
      <xsl:when test="$planet = 'Mars' and $signIdx = 9">true</xsl:when>      <!-- Capricorn -->
      <xsl:when test="$planet = 'Mercury' and $signIdx = 5">true</xsl:when>   <!-- Virgo -->
      <xsl:when test="$planet = 'Jupiter' and $signIdx = 3">true</xsl:when>   <!-- Cancer -->
      <xsl:when test="$planet = 'Venus' and $signIdx = 11">true</xsl:when>    <!-- Pisces -->
      <xsl:when test="$planet = 'Saturn' and $signIdx = 6">true</xsl:when>    <!-- Libra -->
      <xsl:when test="$planet = 'Rahu' and $signIdx = 1">true</xsl:when>      <!-- Taurus (some say Gemini) -->
      <xsl:when test="$planet = 'Ketu' and $signIdx = 7">true</xsl:when>      <!-- Scorpio (some say Sagittarius) -->
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Vedic debilitation (neecha) ───────────────────────────────────── -->
  <xsl:template name="checkVedicDebilitation">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:choose>
      <xsl:when test="$planet = 'Sun' and $signIdx = 6">true</xsl:when>       <!-- Libra -->
      <xsl:when test="$planet = 'Moon' and $signIdx = 7">true</xsl:when>      <!-- Scorpio -->
      <xsl:when test="$planet = 'Mars' and $signIdx = 3">true</xsl:when>      <!-- Cancer -->
      <xsl:when test="$planet = 'Mercury' and $signIdx = 11">true</xsl:when>  <!-- Pisces -->
      <xsl:when test="$planet = 'Jupiter' and $signIdx = 9">true</xsl:when>   <!-- Capricorn -->
      <xsl:when test="$planet = 'Venus' and $signIdx = 5">true</xsl:when>     <!-- Virgo -->
      <xsl:when test="$planet = 'Saturn' and $signIdx = 0">true</xsl:when>    <!-- Aries -->
      <xsl:when test="$planet = 'Rahu' and $signIdx = 7">true</xsl:when>      <!-- Scorpio -->
      <xsl:when test="$planet = 'Ketu' and $signIdx = 1">true</xsl:when>      <!-- Taurus -->
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Vedic state determination ────────────────────────────────────── -->
  <xsl:template name="determineVedicState">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:variable name="swakshetra">
      <xsl:call-template name="checkVedicDomicile">
        <xsl:with-param name="planet" select="$planet"/>
        <xsl:with-param name="signIdx" select="$signIdx"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="uchcha">
      <xsl:call-template name="checkVedicExaltation">
        <xsl:with-param name="planet" select="$planet"/>
        <xsl:with-param name="signIdx" select="$signIdx"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="neecha">
      <xsl:call-template name="checkVedicDebilitation">
        <xsl:with-param name="planet" select="$planet"/>
        <xsl:with-param name="signIdx" select="$signIdx"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:choose>
      <xsl:when test="$swakshetra = 'true'">swakshetra</xsl:when>
      <xsl:when test="$uchcha = 'true'">uchcha</xsl:when>
      <xsl:when test="$neecha = 'true'">neecha</xsl:when>
      <xsl:otherwise>peregrine</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       TITHI COMPUTATION
       ════════════════════════════════════════════════════════════════════ -->

  <xsl:template name="computeTithi">
    <xsl:variable name="moonLon" select="Positions/Planet[@name='Moon']/Sidereal/Lon"/>
    <xsl:variable name="sunLon" select="Positions/Planet[@name='Sun']/Sidereal/Lon"/>
    <xsl:variable name="separation" select="($moonLon - $sunLon + 360) mod 360"/>
    <xsl:variable name="tithiNum" select="floor($separation div 12) + 1"/>
    <xsl:variable name="paksha">
      <xsl:choose>
        <xsl:when test="$separation &lt;= 180">Shukla</xsl:when>
        <xsl:otherwise>Krishna</xsl:otherwise>
      </xsl:choose>
    </xsl:variable>

    <vedic:Number><xsl:value-of select="$tithiNum"/></vedic:Number>
    <vedic:Name>
      <xsl:call-template name="nthToken">
        <xsl:with-param name="list" select="$tithis"/>
        <xsl:with-param name="n" select="$tithiNum"/>
      </xsl:call-template>
    </vedic:Name>
    <vedic:Paksha><xsl:value-of select="$paksha"/></vedic:Paksha>
    <vedic:Separation><xsl:value-of select="format-number($separation, '0.00')"/></vedic:Separation>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       VIMSHOTTARI DASHA
       ════════════════════════════════════════════════════════════════════ -->

  <xsl:template name="computeVimshottari">
    <xsl:variable name="moonLon" select="Positions/Planet[@name='Moon']/Sidereal/Lon"/>
    <xsl:variable name="nakIdx" select="floor($moonLon div 13.333333)"/>
    <xsl:variable name="nakDeg" select="$moonLon - ($nakIdx * 13.333333)"/>
    <xsl:variable name="lordIdx" select="$nakIdx mod 9"/>

    <vedic:MoonNakshatra>
      <xsl:call-template name="nthToken">
        <xsl:with-param name="list" select="$nakshatras"/>
        <xsl:with-param name="n" select="$nakIdx + 1"/>
      </xsl:call-template>
    </vedic:MoonNakshatra>

    <vedic:StartingLord>
      <xsl:call-template name="nthToken">
        <xsl:with-param name="list" select="$nakLords"/>
        <xsl:with-param name="n" select="$lordIdx + 1"/>
      </xsl:call-template>
    </vedic:StartingLord>

    <xsl:variable name="lordYearsStr">
      <xsl:call-template name="nthToken">
        <xsl:with-param name="list" select="$dashaYears"/>
        <xsl:with-param name="n" select="$lordIdx + 1"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="lordYears" select="number($lordYearsStr)"/>

    <vedic:LordYears><xsl:value-of select="$lordYears"/></vedic:LordYears>
    <vedic:ElapsedInNakshatra><xsl:value-of select="format-number($nakDeg, '0.00')"/></vedic:ElapsedInNakshatra>
    <vedic:RemainingYears>
      <xsl:value-of select="format-number((1 - ($nakDeg div 13.333333)) * $lordYears, '0.00')"/>
    </vedic:RemainingYears>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       NODES (RAHU / KETU)
       ════════════════════════════════════════════════════════════════════ -->

  <xsl:template name="computeNodes">
    <xsl:variable name="rahuLon" select="Positions/Planet[@name='TrueNode']/Sidereal/Lon"/>
    <xsl:variable name="ketuLon" select="($rahuLon + 180) mod 360"/>

    <vedic:Rahu>
      <vedic:Sign>
        <xsl:call-template name="lonToSign">
          <xsl:with-param name="lon" select="$rahuLon"/>
        </xsl:call-template>
      </vedic:Sign>
      <vedic:Degree>
        <xsl:value-of select="format-number($rahuLon mod 30, '0.00')"/>
      </vedic:Degree>
    </vedic:Rahu>

    <vedic:Ketu>
      <vedic:Sign>
        <xsl:call-template name="lonToSign">
          <xsl:with-param name="lon" select="$ketuLon"/>
        </xsl:call-template>
      </vedic:Sign>
      <vedic:Degree>
        <xsl:value-of select="format-number($ketuLon mod 30, '0.00')"/>
      </vedic:Degree>
    </vedic:Ketu>
  </xsl:template>

</xsl:stylesheet>
